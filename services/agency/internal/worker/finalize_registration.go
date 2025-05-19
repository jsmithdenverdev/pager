package worker

import (
	"context"
	"encoding/json"
	"fmt"
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

		if _, err := dynamoClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName: aws.String(config.AgencyTableName),
			Key: map[string]types.AttributeValue{
				"pk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("agency#%s", message.AgencyID),
				},
				"sk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("registration#%s", message.RegistrationCode),
				},
			},
			UpdateExpression: aws.String("set #status = :status"),
			ExpressionAttributeNames: map[string]string{
				"#status": "status",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":status": &types.AttributeValueMemberS{
					Value: "ACTIVE",
				},
			},
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
