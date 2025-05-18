package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/auth0/go-auth0/authentication"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/caarlos0/env/v11"
	"github.com/jsmithdenverdev/pager/services/user/internal/worker"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "run failed: %s", err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	var conf worker.Config
	if err := env.Parse(&conf); err != nil {
		return fmt.Errorf("failed to load config from env: %w", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.Level(conf.LogLevel),
	}))

	awsconf, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load default aws config: %w", err)
	}

	dynamoClient := dynamodb.NewFromConfig(awsconf)
	snsClient := sns.NewFromConfig(awsconf)

	// Initialize a new client using a domain, client ID and client secret.
	authAPI, err := authentication.New(
		ctx,
		conf.Auth0Domain,
		authentication.WithClientID(conf.Auth0ManagementClientID),
		authentication.WithClientSecret(conf.Auth0ManagementClientSecret),
	)
	if err != nil {
		return fmt.Errorf("failed to initialize the auth0 authentication API client: %w", err)
	}

	lambda.Start(worker.ProcessEvents(conf, logger, dynamoClient, snsClient, authAPI))

	return nil
}
