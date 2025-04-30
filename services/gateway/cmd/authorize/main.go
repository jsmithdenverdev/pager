package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/caarlos0/env/v10"
	"github.com/jsmithdenverdev/pager/services/gateway/internal/app"
)

func main() {
	if err := run(context.Background(), os.Stdout, os.Getenv); err != nil {
		fmt.Fprintf(os.Stderr, "run failed: %s", err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context, stdout io.Writer, getenv func(string) string) error {

	conf := app.Config{}
	if err := env.Parse(&conf); err != nil {
		return fmt.Errorf("[in main.run] failed to load config from env: %w", err)
	}

	logger := slog.New(slog.NewJSONHandler(stdout, &slog.HandlerOptions{
		Level: conf.LogLevel,
	}))

	awsconf, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("[in main.run] failed to load default aws config: %w", err)
	}

	dynamodb := dynamodb.NewFromConfig(awsconf)

	lambda.StartWithOptions(app.Authorize(conf, logger, dynamodb))

	return nil
}
