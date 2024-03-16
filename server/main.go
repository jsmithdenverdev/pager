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

	"github.com/go-playground/validator/v10"
	"github.com/graphql-go/handler"
)

func main() {
	if err := run(context.Background(), os.Stdout, os.Getenv); err != nil {
		fmt.Fprint(os.Stderr, err.Error())
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

	validate := validator.New(validator.WithRequiredStructEnabled())

	schema, err := newSchema(logger, validate)
	if err != nil {
		return err
	}

	handler := handler.New(&handler.Config{
		Schema: &schema,
		Pretty: true,
	})

	mux := http.NewServeMux()
	mux.Handle("/graphql", handler)

	httpServer := http.Server{
		Addr:    net.JoinHostPort(getenv("HOST"), getenv("PORT")),
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
