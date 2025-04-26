package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/a-h/awsapigatewayv2handler"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/caarlos0/env/v11"
	"github.com/jsmithdenverdev/pager/services/agency/internal/app"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "run failed: %s", err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	var conf app.Config
	if err := env.Parse(&conf); err != nil {
		return fmt.Errorf("failed to load config from env: %w", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	handler := app.NewServer(conf, logger)

	lambda.Start(awsapigatewayv2handler.NewLambdaHandler(handler))

	return nil
}
