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
	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
	"github.com/jsmithdenverdev/pager/pkg/middleware/apigateway"
	"github.com/jsmithdenverdev/pager/services/page/internal/config"
	"github.com/jsmithdenverdev/pager/services/page/internal/handlers"
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
	fmt.Fprintf(stdout, "Version %s", Version)

	conf, err := config.LoadFromEnv(getenv)
	if err != nil {
		return fmt.Errorf("[in main.run] failed to load config from env: %w", err)
	}

	logger := slog.New(slog.NewJSONHandler(stdout, nil))

	awsconf, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("[in main.run] failed to load default aws config: %w", err)
	}

	verifiedPermissionsClient := verifiedpermissions.NewFromConfig(awsconf)

	dynamodbClient := dynamodb.NewFromConfig(awsconf)

	handler := handlers.CreateAgency(
		conf,
		logger,
		dynamodbClient,
	)

	handler = apigateway.WithAuthz(conf.PolicyStoreID, verifiedPermissionsClient, logger)(handler)

	lambda.StartWithOptions(handler)

	return nil
}
