package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/jsmithdenverdev/pager/services/agency/internal/models"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
)

// finalizeRegistration finalizes an endpoint registration.
func finalizeRegistration(config Config, logger *slog.Logger, dynamoClient *dynamodb.Client, snsClient *sns.Client) func(context.Context, events.SNSEntity, int) error {
	type message struct {
		RegistrationCode string `json:"registrationCode"`
		AgencyID         string `json:"agencyId"`
		EndpointId       string `json:"endpointId"`
	}

	logAndHandleError := eventProcessorErrorHandler(config, logger, snsClient, evtRegistrationCreateFailed)

	return func(ctx context.Context, record events.SNSEntity, retryCount int) error {
		var message message

		if err := json.Unmarshal([]byte(record.Message), &message); err != nil {
			return logAndHandleError(ctx, retryCount, "failed to create registration", message, err)
		}

		readPendingRegistrationResult, err := dynamoClient.GetItem(ctx, &dynamodb.GetItemInput{
			TableName: aws.String(config.AgencyTableName),
			Key: map[string]types.AttributeValue{
				"pk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("agency#%s", message.AgencyID),
				},
				"sk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("registration#%s", message.RegistrationCode),
				},
			},
		})

		if err != nil {
			return logAndHandleError(ctx, retryCount, "failed to create registration", message, err)
		}

		if readPendingRegistrationResult.Item == nil {
			return logAndHandleError(ctx, retryCount, "failed to create registration", message, errors.New("registration doesn't exist"))
		}

		var pendingRegistration models.EndpointRegistration
		if err := attributevalue.UnmarshalMap(readPendingRegistrationResult.Item, &pendingRegistration); err != nil {
			return logAndHandleError(ctx, retryCount, "failed to create registration", message, err)
		}

		if pendingRegistration.Status != models.RegistrationStatusPending {
			return logAndHandleError(ctx, retryCount, "failed to create registration", message, errors.New("registration is not pending"))
		}

		// Mirror the pending registration into a new finalRegistration model to avoid any mutation of the original.
		finalRegistration := pendingRegistration
		finalRegistration.Status = models.RegistrationStatusComplete
		finalRegistration.SK = fmt.Sprintf("endpoint#%s", message.EndpointId)

		finalRegistrationAV, err := attributevalue.MarshalMap(finalRegistration)
		if err != nil {
			return logAndHandleError(ctx, retryCount, "failed to create registration", message, err)
		}

		// When we finalize a registration we replace registrationcode sort key on the record with an endpoint
		// identifier. Sort keys can't be updated, so we need to delete the record and create a new one.
		// I'd prefer to do this in a dynamodb transaction but you can't perform both a delete and put in a single
		// transaction.
		if _, err := dynamoClient.DeleteItem(ctx, &dynamodb.DeleteItemInput{
			TableName: aws.String(config.AgencyTableName),
			Key: map[string]types.AttributeValue{
				"pk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("agency#%s", message.AgencyID),
				},
				"sk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("registration#%s", message.RegistrationCode),
				},
			},
		}); err != nil {
			return logAndHandleError(ctx, retryCount, "failed to create registration", message, err)
		}

		if _, err := dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(config.AgencyTableName),
			Item:      finalRegistrationAV,
		}); err != nil {
			return logAndHandleError(ctx, retryCount, "failed to create registration", message, err)
		}

		if _, err := snsClient.Publish(ctx, &sns.PublishInput{
			TopicArn: aws.String(config.EventsTopicARN),
			Message:  aws.String(fmt.Sprintf(`{"endpointId": "%s", "agencyId": "%s"}`, message.EndpointId, message.AgencyID)),
			MessageAttributes: map[string]snstypes.MessageAttributeValue{
				"type": {
					DataType:    aws.String("String"),
					StringValue: aws.String(evtRegistrationCreated),
				},
			},
		}); err != nil {
			return logAndHandleError(ctx, retryCount, "failed to create registration", message, err)
		}

		logger.DebugContext(ctx, "published event", slog.String("type", evtRegistrationCreated))

		return nil
	}
}
