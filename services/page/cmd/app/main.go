package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/a-h/awsapigatewayv2handler"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "run failed: %s", err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	loghandler := slog.NewJSONHandler(os.Stdout, nil)
	lambda.Start(awsapigatewayv2handler.NewLambdaHandler(newServer(loghandler)))
	return nil
}

func newServer(loghandler slog.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		logger := slog.New(loghandler)
		logger.InfoContext(r.Context(), "request received", "event", r)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("list pages"))
	})

	mux.HandleFunc("GET /{id}", func(w http.ResponseWriter, r *http.Request) {
		logger := slog.New(loghandler)
		logger.InfoContext(r.Context(), "request received", "event", r)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("read page by id"))
	})

	mux.HandleFunc("GET /agencies/{agencyid}/pages", func(w http.ResponseWriter, r *http.Request) {
		logger := slog.New(loghandler)
		logger.InfoContext(r.Context(), "request received", "event", r)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("list pages by agency"))
	})

	mux.HandleFunc("GET /agencies/{agencyid}/pages/{pageid}", func(w http.ResponseWriter, r *http.Request) {
		logger := slog.New(loghandler)
		logger.InfoContext(r.Context(), "request received", "event", r)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("read page by agency and id"))
	})

	return mux
}
