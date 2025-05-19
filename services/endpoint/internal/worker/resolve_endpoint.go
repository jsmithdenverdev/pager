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

func resolveEndpoint(config Config, logger *slog.Logger, dynamoClient *dynamodb.Client, snsClient *sns.Client) func(ctx context.Context, snsRecord events.SNSEntity, retryCount int) error {
	type message struct {
		AgencyID         string `json:"agencyId"`
		RegistrationCode string `json:"registrationCode"`
	}

	logAndHandleError := eventProcessorErrorHandler(config, logger, snsClient, evtEndpointResolutionFailed)

	return func(ctx context.Context, snsRecord events.SNSEntity, retryCount int) error {
		var message message

		if err := json.Unmarshal([]byte(snsRecord.Message), &message); err != nil {
			return logAndHandleError(ctx, retryCount, "failed to resolve endpoint from registration code", message, err)
		}

		queryRegistrationCodeResult, err := dynamoClient.Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(config.EndpointTableName),
			KeyConditionExpression: aws.String("#pk = :pk AND #sk = :sk"),
			ExpressionAttributeNames: map[string]string{
				"#pk": "pk",
				"#sk": "sk",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("rc#%s", message.RegistrationCode),
				},
				":sk": &types.AttributeValueMemberS{
					Value: "registrationcode",
				},
			},
		})

		if err != nil {
			return logAndHandleError(ctx, retryCount, "failed to resolve endpoint from registration code", message, err)
		}

		if len(queryRegistrationCodeResult.Items) == 0 {
			return logAndHandleError(ctx, retryCount, "failed to resolve endpoint from registration code", message, errors.New("registration code doesn't exist"))
		}

		var rc models.RegistrationCode

		if err := attributevalue.UnmarshalMap(queryRegistrationCodeResult.Items[0], &rc); err != nil {
			return logAndHandleError(ctx, retryCount, "failed to resolve endpoint from registration code", message, err)
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
					Value: fmt.Sprintf("endpoint#%s", rc.EndpointID),
				},
				":sk": &types.AttributeValueMemberS{
					Value: "meta",
				},
			},
		})

		if err != nil {
			return logAndHandleError(ctx, retryCount, "failed to resolve endpoint from registration code", message, err)
		}

		if len(queryEndpointResult.Items) == 0 {
			return logAndHandleError(ctx, retryCount, "failed to resolve endpoint from registration code", message, errors.New("endpoint doesn't exist"))
		}

		var endpoint models.Endpoint

		if err := attributevalue.UnmarshalMap(queryEndpointResult.Items[0], &endpoint); err != nil {
			return logAndHandleError(ctx, retryCount, "failed to resolve endpoint from registration code", message, err)
		}

		messageBody, err := json.Marshal(struct {
			RegistrationCode string `json:"registrationCode"`
			AgencyID         string `json:"agencyId"`
			EndpointID       string `json:"endpointId"`
		}{
			RegistrationCode: message.RegistrationCode,
			AgencyID:         message.AgencyID,
			EndpointID:       rc.EndpointID,
		})

		if err != nil {
			return logAndHandleError(ctx, retryCount, "failed to resolve endpoint from registration code", message, err)
		}

		if _, err := snsClient.Publish(ctx, &sns.PublishInput{
			TopicArn: aws.String(config.EventsTopicARN),
			Message:  aws.String(string(messageBody)),
			MessageAttributes: map[string]snstypes.MessageAttributeValue{
				"type": {
					DataType:    aws.String("String"),
					StringValue: aws.String(evtEndpointResolved),
				},
			},
		}); err != nil {
			return logAndHandleError(ctx, retryCount, "failed to resolve endpoint from registration code", message, err)
		}

		return nil
	}
}
