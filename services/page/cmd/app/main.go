package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/events"
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
	lambda.Start(handler(loghandler))
	return nil
}

func handler(loghandler slog.Handler) func(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return func(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		logger := slog.New(loghandler)
		logger.InfoContext(ctx, "request received", "event", event)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 200,
		}, nil
	}
}
