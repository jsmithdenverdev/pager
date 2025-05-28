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
)

// finalizeRegistration finalizes an endpoint registration.
func markRegistrationFailed(config Config, logger *slog.Logger, dynamoClient *dynamodb.Client) func(context.Context, events.SNSEntity, int) error {
	type message struct {
		RegistrationCode string `json:"registrationCode"`
		AgencyID         string `json:"agencyId"`
		EndpointId       string `json:"userId"`
	}

	return func(ctx context.Context, record events.SNSEntity, retryCount int) error {
		var message message

		if err := json.Unmarshal([]byte(record.Message), &message); err != nil {
			logger.ErrorContext(ctx, "failed to unmarshal message", slog.Any("error", err))
			return err
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
					Value: "FAILED",
				},
			},
		}); err != nil {
			logger.ErrorContext(ctx, "failed to update registration", slog.Any("error", err))
			return err
		}

		return nil
	}
}
