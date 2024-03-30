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

	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
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

	config, err := newConfigFromProcessEnv(getenv)
	if err != nil {
		return err
	}

	validate := validator.New(validator.WithRequiredStructEnabled())

	authz, err := authzed.NewClient(
		config.SpiceDBEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpcutil.WithInsecureBearerToken(config.SpiceDBToken))

	if err != nil {
		return err
	}

	db, err := sqlx.Connect("postgres", config.DBConn)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	addRoutes(mux, config, logger, validate, authz, db)

	httpServer := http.Server{
		Addr:    net.JoinHostPort(config.Host, config.Port),
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
