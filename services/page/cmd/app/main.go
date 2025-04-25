package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
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

func handler(loghandler slog.Handler) func(context.Context, events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	logger := slog.New(loghandler)
	return func(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		logger.InfoContext(ctx, "pages service request received", slog.Any("request", request))
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusOK,
		}, nil
	}
}
