package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
)

// CreateMembership creates a new membership in the agency.
func deleteMembership(config Config, logger *slog.Logger, dynamoClient *dynamodb.Client, snsClient *sns.Client) func(context.Context, events.SNSEntity, int) error {
	type message struct {
		UserID   string `json:"userId"`
		AgencyID string `json:"agencyId"`
	}

	logAndHandleError := eventProcessorErrorHandler(config, logger, snsClient, evtMembershipDeleteFailed)

	return func(ctx context.Context, record events.SNSEntity, retryCount int) error {
		var message message

		if err := json.Unmarshal([]byte(record.Message), &message); err != nil {
			return logAndHandleError(ctx, retryCount, "failed to delete membership", message, err)
		}

		// Add condition expressions to ensure items exist before deletion
		// This will make the transaction fail if either item doesn't exist
		_, err := dynamoClient.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
			TransactItems: []types.TransactWriteItem{
				{
					Delete: &types.Delete{
						Key: map[string]types.AttributeValue{
							"pk": &types.AttributeValueMemberS{
								Value: fmt.Sprintf("user#%s", message.UserID),
							},
							"sk": &types.AttributeValueMemberS{
								Value: fmt.Sprintf("agency#%s", message.AgencyID),
							},
						},
						TableName:           aws.String(config.AgencyTableName),
						ConditionExpression: aws.String("attribute_exists(pk) AND attribute_exists(sk)"),
					},
				},
				{
					Delete: &types.Delete{
						Key: map[string]types.AttributeValue{
							"pk": &types.AttributeValueMemberS{
								Value: fmt.Sprintf("agency#%s", message.AgencyID),
							},
							"sk": &types.AttributeValueMemberS{
								Value: fmt.Sprintf("user#%s", message.UserID),
							},
						},
						TableName:           aws.String(config.AgencyTableName),
						ConditionExpression: aws.String("attribute_exists(pk) AND attribute_exists(sk)"),
					},
				},
			},
		})

		if err != nil {
			// Check if the error is a condition check failure, which means records didn't exist
			var conditionCheckFailed *types.TransactionCanceledException
			if !errors.As(err, &conditionCheckFailed) {
				// This is a different error, not just that the records didn't exist
				return logAndHandleError(ctx, retryCount, "failed to delete membership", message, err)
			}

			// If we get here, it means the records didn't exist, so we don't publish an event
			logger.Info("membership records not found, skipping event publication",
				"userId", message.UserID,
				"agencyId", message.AgencyID)
			return nil
		}

		// If we get here, the delete was successful, so publish the event
		if _, err := snsClient.Publish(ctx, &sns.PublishInput{
			TopicArn: aws.String(config.EventsTopicARN),
			Message:  aws.String(fmt.Sprintf(`{"userId": "%s", "agencyId": "%s"}`, message.UserID, message.AgencyID)),
			MessageAttributes: map[string]snstypes.MessageAttributeValue{
				"type": {
					DataType:    aws.String("String"),
					StringValue: aws.String(evtMembershipDeleted),
				},
			},
		}); err != nil {
			return logAndHandleError(ctx, retryCount, "failed to publish delete membership event", message, err)
		}

		return nil
	}
}
