package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/jsmithdenverdev/pager/services/endpoint/internal/models"
)

func upsertRegistration(config Config, logger *slog.Logger, dynamoClient *dynamodb.Client, snsClient *sns.Client) func(ctx context.Context, snsRecord events.SNSEntity, retryCount int) error {
	type message struct {
		AgencyID   string `json:"agencyId"`
		EndpointID string `json:"endpointId"`
	}

	logAndHandleError := eventProcessorErrorHandler(config, logger, snsClient, evtRegistrationUpsertFailed)

	return func(ctx context.Context, snsRecord events.SNSEntity, retryCount int) error {
		var message message

		if err := json.Unmarshal([]byte(snsRecord.Message), &message); err != nil {
			return logAndHandleError(ctx, retryCount, "failed to upsert endpoint registration", message, err)
		}

		queryEndpointResult, err := dynamoClient.Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(config.EndpointTableName),
			KeyConditionExpression: aws.String("#pk = :pk AND #sk = :sk"),
			ExpressionAttributeNames: map[string]string{
				"#pk": "pk",
				"#sk": "sk",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("endpoint#%s", message.EndpointID),
				},
				":sk": &types.AttributeValueMemberS{
					Value: "meta",
				},
			},
		})

		if err != nil {
			return logAndHandleError(ctx, retryCount, "failed to upsert endpoint registration", message, err)
		}

		if len(queryEndpointResult.Items) == 0 {
			return logAndHandleError(ctx, retryCount, "failed to upsert endpoint registration", message, errors.New("endpoint doesn't exist"))
		}

		var endpoint models.Endpoint

		if err := attributevalue.UnmarshalMap(queryEndpointResult.Items[0], &endpoint); err != nil {
			return logAndHandleError(ctx, retryCount, "failed to upsert endpoint registration", message, err)
		}

		if endpoint.Registrations == nil {
			endpoint.Registrations = map[string]any{}
		}
		endpoint.Registrations[message.AgencyID] = struct{}{}

		registrations, err := attributevalue.MarshalMap(endpoint.Registrations)
		if err != nil {
			return logAndHandleError(ctx, retryCount, "failed to upsert endpoint registration", message, err)
		}

		if _, err := dynamoClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName: aws.String(config.EndpointTableName),
			Key: map[string]types.AttributeValue{
				"pk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("endpoint#%s", message.EndpointID),
				},
				"sk": &types.AttributeValueMemberS{
					Value: "meta",
				},
			},
			UpdateExpression: aws.String("SET #registrations = :registrations"),
			ExpressionAttributeNames: map[string]string{
				"#registrations": "registrations",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":registrations": &types.AttributeValueMemberM{
					Value: registrations,
				},
			},
		}); err != nil {
			return logAndHandleError(ctx, retryCount, "failed to upsert endpoint registration", message, err)
		}

		messageBody, err := json.Marshal(struct {
			AgencyID   string `json:"agencyId"`
			EndpointID string `json:"endpointId"`
		}{
			AgencyID:   message.AgencyID,
			EndpointID: message.EndpointID,
		})

		if err != nil {
			return logAndHandleError(ctx, retryCount, "failed to upsert endpoint registration", message, err)
		}

		if _, err = snsClient.Publish(ctx, &sns.PublishInput{
			TopicArn: aws.String(config.EventsTopicARN),
			Message:  aws.String(string(messageBody)),
			MessageAttributes: map[string]snstypes.MessageAttributeValue{
				"type": {
					DataType:    aws.String("String"),
					StringValue: aws.String(evtRegistrationUpserted),
				},
			},
		}); err != nil {
			return logAndHandleError(ctx, retryCount, "failed to upsert endpoint registration", message, err)
		}

		return nil
	}
}
