package app

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// CreateMembership creates a new membership in the agency.
func createMembership(config Config, logger *slog.Logger, dynamoClient *dynamodb.Client, snsClient *sns.Client) func(context.Context, events.SQSMessage) error {
	return func(ctx context.Context, record events.SQSMessage) error {
		var body any
		if err := json.Unmarshal([]byte(record.Body), &body); err != nil {
			return err
		}
		logger.DebugContext(ctx, "create membership", slog.Any("body", body))
		return nil
	}
}
