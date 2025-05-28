package worker

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"log/slog"
)

func trackFailedDelivery(config Config, logger *slog.Logger, dynamoClient *dynamodb.Client) func(context.Context, events.SNSEntity, int) error {
	return func(ctx context.Context, entity events.SNSEntity, i int) error {
		return nil
	}
}
