package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/a-h/awsapigatewayv2handler"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/caarlos0/env/v11"
)

type Config struct {
	Environment string `env:"ENVIRONMENT"`
}

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "run failed: %s", err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	var conf Config
	if err := env.Parse(&conf); err != nil {
		return fmt.Errorf("[in main.run] failed to load config from env: %w", err)
	}

	loghandler := slog.NewJSONHandler(os.Stdout, nil)
	lambda.Start(awsapigatewayv2handler.NewLambdaHandler(newServer(conf, loghandler)))
	return nil
}

func newServer(config Config, loghandler slog.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc(fmt.Sprintf("GET /%s", config.Environment), func(w http.ResponseWriter, r *http.Request) {
		logger := slog.New(loghandler)
		logger.InfoContext(r.Context(), "request received", slog.Any("request.headers", r.Header))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("list agencies"))
	})

	mux.HandleFunc(fmt.Sprintf("GET /%s/{id}", config.Environment), func(w http.ResponseWriter, r *http.Request) {
		logger := slog.New(loghandler)
		logger.InfoContext(r.Context(), "request received", slog.Any("request.headers", r.Header))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("read agency by id"))
	})

	return mux
}
