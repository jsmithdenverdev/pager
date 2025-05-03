package app

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

func EventProcessor(config Config, logger *slog.Logger, dynamoClient *dynamodb.Client, snsClient *sns.Client) func(ctx context.Context, event events.SQSEvent) (events.SQSEventResponse, error) {
	return func(ctx context.Context, event events.SQSEvent) (events.SQSEventResponse, error) {
		var batchItemFailures []events.SQSBatchItemFailure
		for _, record := range event.Records {
			logger.DebugContext(ctx, "processing record", slog.Any("record", record))
			// Unmarshal the record body into a SNSEntity
			var snsRecord events.SNSEntity
			if err := json.Unmarshal([]byte(record.Body), &snsRecord); err != nil {
				logger.ErrorContext(ctx, "failed to unmarshal record body", slog.Any("error", err))
				batchItemFailures = append(batchItemFailures, events.SQSBatchItemFailure{
					ItemIdentifier: record.MessageId,
				})
				continue
			}

			eventType := snsRecord.MessageAttributes["type"].(map[string]any)["Value"].(string)
			// Use a type attribute on the message to determine the event type
			switch eventType {
			case "user.user.invite":
				if err := inviteUser(config, logger, dynamoClient, snsClient)(ctx, snsRecord); err != nil {
					logger.ErrorContext(ctx, "failed to invite user", slog.Any("error", err))
					batchItemFailures = append(batchItemFailures, events.SQSBatchItemFailure{
						ItemIdentifier: record.MessageId,
					})
				}
			default:
				logger.ErrorContext(
					ctx,
					"unknown event type",
					slog.Any("type", snsRecord.MessageAttributes["type"]),
					slog.String("messageId", record.MessageId))

				batchItemFailures = append(batchItemFailures, events.SQSBatchItemFailure{
					ItemIdentifier: record.MessageId,
				})
			}
		}

		return events.SQSEventResponse{
			BatchItemFailures: batchItemFailures,
		}, nil
	}
}
