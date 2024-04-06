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

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	jwtvalidator "github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"github.com/graphql-go/handler"
	"github.com/jmoiron/sqlx"
	"github.com/jsmithdenverdev/pager/authz"
	"github.com/jsmithdenverdev/pager/config"
	"github.com/jsmithdenverdev/pager/middleware"
	"github.com/jsmithdenverdev/pager/schema"
	"github.com/jsmithdenverdev/pager/service"
	_ "github.com/lib/pq"
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

	schema, err := schema.New()
	if err != nil {
		return err
	}

	handler := handler.New(&handler.Config{
		Schema: &schema,
		Pretty: true,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.Handle("/graphql", middleware.EnsureValidToken(cfg, logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims := ctx.Value(jwtmiddleware.ContextKey{}).(*jwtvalidator.ValidatedClaims)

		authz := authz.NewClient(ctx, authzedClient, logger, claims.RegisteredClaims.Subject)
		userService := service.NewUserService(ctx, db, claims.RegisteredClaims.Subject)
		agencyService := service.NewAgencyService(ctx, authz, db, logger)

		ctx = context.WithValue(ctx, service.ContextKeyUserService, userService)
		ctx = context.WithValue(ctx, service.ContextKeyAgencyService, agencyService)

		handler.ContextHandler(ctx, w, r)
	})))

	httpServer := http.Server{
		Addr:    net.JoinHostPort(cfg.Host, cfg.Port),
		Handler: mux,
	}

	go func() {
		logger.InfoContext(ctx, fmt.Sprintf("listening for connections on %s", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil {
			logger.ErrorContext(ctx, fmt.Sprintf("error listening and serving %s", err.Error()))
		}
	}()

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
