package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	if err := run(context.Background(), os.Stdout); err != nil {
		fmt.Fprint(os.Stderr, "run failed: %s", err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context, stdout io.Writer) error {
	logger := slog.New(slog.NewJSONHandler(stdout, nil))
	lambda.Start(handler(logger))
	return nil
}

func handler(logger *slog.Logger) func(ctx context.Context, event events.SQSEvent) (events.SQSEventResponse, error) {
	return func(ctx context.Context, event events.SQSEvent) (events.SQSEventResponse, error) {
		logger.InfoContext(ctx, "user event processor", slog.Any("event", event))
		return events.SQSEventResponse{}, nil
	}
}
