package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
)

const (
	evtEndpointResolved         = "endpoint.resolved"
	evtEndpointResolutionFailed = "endpoint.resolution.failed"
	evtRegistrationUpserted     = "endpoint.registration.upserted"
	evtRegistrationUpsertFailed = "endpoint.registration.upsert.failed"
	evtRegistrationDeleted      = "endpoint.registration.deleted"
	evtRegistrationDeleteFailed = "endpoint.registration.delete.failed"
)

func EventProcessor(config Config, logger *slog.Logger, dynamoClient *dynamodb.Client, snsClient *sns.Client) func(ctx context.Context, event events.SQSEvent) (events.SQSEventResponse, error) {
	return func(ctx context.Context, event events.SQSEvent) (events.SQSEventResponse, error) {
		var batchItemFailures []events.SQSBatchItemFailure
		for _, record := range event.Records {
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
			receiveCount, err := strconv.Atoi(record.Attributes["ApproximateReceiveCount"])
			if err != nil {
				logger.ErrorContext(ctx, "failed to convert receive count to int", slog.Any("error", err))
				batchItemFailures = append(batchItemFailures, events.SQSBatchItemFailure{
					ItemIdentifier: record.MessageId,
				})
				continue
			}
			retryCount := receiveCount + 1
			// Use a type attribute on the message to determine the event type
			switch eventType {
			case "endpoint.resolve":
				if err := resolveEndpoint(config, logger, dynamoClient, snsClient)(ctx, snsRecord, retryCount); err != nil {
					logger.ErrorContext(ctx, "failed to resolve registration", slog.Any("error", err))
					batchItemFailures = append(batchItemFailures, events.SQSBatchItemFailure{
						ItemIdentifier: record.MessageId,
					})
				}
			case "endpoint.deliver":
				if err := deliverToEndpoints(config, logger, dynamoClient, snsClient)(ctx, snsRecord, retryCount); err != nil {
					logger.ErrorContext(ctx, "failed to deliver to endpoints", slog.Any("error", err))
					batchItemFailures = append(batchItemFailures, events.SQSBatchItemFailure{
						ItemIdentifier: record.MessageId,
					})
				}
			case "agency.registration.created", "agency.registration.updated":
				if err := upsertRegistration(config, logger, dynamoClient, snsClient)(ctx, snsRecord, retryCount); err != nil {
					logger.ErrorContext(ctx, "failed to upsert registration", slog.Any("error", err))
					batchItemFailures = append(batchItemFailures, events.SQSBatchItemFailure{
						ItemIdentifier: record.MessageId,
					})
				}
			case "agency.registration.deleted":
				if err := deleteRegistration(config, logger, dynamoClient, snsClient)(ctx, snsRecord, retryCount); err != nil {
					logger.ErrorContext(ctx, "failed to delete registration", slog.Any("error", err))
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

func eventProcessorErrorHandler(config Config, logger *slog.Logger, snsClient *sns.Client, eventType string) func(ctx context.Context, retryCount int, msg string, event any, err error, attributes ...any) error {
	return func(ctx context.Context, retryCount int, msg string, event any, err error, attributes ...any) error {
		logger.ErrorContext(ctx, msg, append(attributes, slog.Any("error", err))...)
		if retryCount >= config.EventRetryCount {
			messageBytes, err := json.Marshal(event)
			if err != nil {
				logger.ErrorContext(ctx, "failed to marshal event to json", append(attributes, slog.Any("error", err))...)
				return err
			}
			if _, err := snsClient.Publish(ctx, &sns.PublishInput{
				TopicArn: aws.String(config.EventsTopicARN),
				Message:  aws.String(string(messageBytes)),
				MessageAttributes: map[string]snstypes.MessageAttributeValue{
					"type": {
						DataType:    aws.String("String"),
						StringValue: aws.String(eventType),
					},
				},
			}); err != nil {
				logger.ErrorContext(ctx, "failed to event", append(attributes, slog.Any("error", err))...)
				return err
			}
		}
		return err
	}
}
