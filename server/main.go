package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/auth0/go-auth0/management"
	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"github.com/go-chi/chi"
	"github.com/jmoiron/sqlx"
	"github.com/jsmithdenverdev/pager/authz"
	"github.com/jsmithdenverdev/pager/config"
	"github.com/jsmithdenverdev/pager/pubsub"
	"github.com/jsmithdenverdev/pager/routes"
	"github.com/jsmithdenverdev/pager/worker"
	pq "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if err := run(context.Background(), os.Stdout, os.Getenv); err != nil {
		fmt.Fprintf(os.Stderr, "error: could not start pager: %s", err.Error())
		os.Exit(1)
	}
}

// run initializes dependencies and kicks this pig.
func run(ctx context.Context, stdout io.Writer, getenv func(string) string) error {
	logger := slog.New(slog.NewJSONHandler(stdout, &slog.HandlerOptions{
		AddSource: true,
	}))

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	cfg, err := config.LoadFromEnv(getenv)
	if err != nil {
		return err
	}

	authzedClient, err := authzed.NewClient(
		cfg.SpiceDBEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpcutil.WithInsecureBearerToken(cfg.SpiceDBToken))
	if err != nil {
		return err
	}

	db, err := sqlx.Connect("postgres", cfg.DBConn)
	if err != nil {
		return err
	}

	ln := pq.NewListener(
		cfg.DBConn,
		time.Duration(5)*time.Second,
		time.Duration(30)*time.Second,
		func(event pq.ListenerEventType, err error) {
			if err != nil {
				logger.ErrorContext(
					ctx,
					"pq listener event callback error",
					"event", event,
					"error", err)
			}
			logger.InfoContext(
				ctx,
				"pq listener event callback",
				"event", event)
		})

	auth0, err := management.New(
		cfg.Auth0Domain,
		management.WithClientCredentials(
			ctx,
			cfg.Auth0ClientID,
			cfg.Auth0ClientSecret,
		))
	if err != nil {
		return err
	}

	pubsubClient := pubsub.NewClient(ctx, db, ln, logger)

	router := chi.NewMux()
	if err := routes.Register(
		router,
		logger,
		cfg,
		authzedClient,
		db,
		auth0,
		pubsubClient); err != nil {
		return fmt.Errorf("failed to register routes: %w", err)
	}

	httpServer := http.Server{
		Addr:    net.JoinHostPort(cfg.Host, cfg.Port),
		Handler: router,
	}

	// Start http server
	go func() {
		logger.InfoContext(ctx, fmt.Sprintf("listening for connections on %s", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil {
			logger.ErrorContext(ctx, fmt.Sprintf("error listening and serving %s", err.Error()))
		}
	}()

	// Subscribe to topics. We publish to a topic by writing to a table. This
	// enables us to treat pubsub transactionally and only publish a message
	// once we've succesfully written data to a table.
	if err := pubsub.Subscribe(
		pubsubClient,
		pubsub.TopicProvisionUser,
		worker.NewProvisionUserHandler(
			ctx,
			db,
			// ProvisionUser only uses WritePermissions which does not need a user id
			authz.NewClient(ctx, authzedClient, logger, ""),
			auth0,
			logger)); err != nil {
		return err
	}
	if err := pubsub.Subscribe(
		pubsubClient,
		pubsub.TopicSendPage,
		worker.NewSendPageHandler(ctx, db, logger)); err != nil {
		return err
	}

	// go pubsubClient.Start(ctx)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
	}()
	wg.Wait()
	return nil
}
