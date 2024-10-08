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
	"github.com/jsmithdenverdev/pager/services/user/internal/config"
	"github.com/jsmithdenverdev/pager/services/user/internal/handlers"
)

var (
	Version string
)

func main() {
	if err := run(context.Background(), os.Stdout, os.Getenv); err != nil {
		fmt.Fprintf(os.Stderr, "run failed: %s", err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context, stdout io.Writer, getenv func(string) string) error {
	logger := slog.New(slog.NewJSONHandler(stdout, nil))

	conf, err := config.LoadFromEnv(getenv)
	if err != nil {
		return fmt.Errorf("[in main.run] failed to load config from env: %w", err)
	}

	awsconf, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("[in main.run] failed to load default aws config: %w", err)
	}

	dynamodb := dynamodb.NewFromConfig(awsconf)

	lambda.StartWithOptions(handlers.UserInfo(conf, logger, dynamodb))

	return nil
}
