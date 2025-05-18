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
)

func markInviteFailed(config Config, logger *slog.Logger, dynamoClient *dynamodb.Client, snsClient *sns.Client) func(context.Context, events.SNSEntity, int) error {
	return func(ctx context.Context, record events.SNSEntity, retryCount int) error {
		var message struct {
			Email    string `json:"email"`
			AgencyID string `json:"agencyId"`
		}

		if err := json.Unmarshal([]byte(record.Message), &message); err != nil {
			logger.ErrorContext(ctx, "failed to unmarshal message", slog.Any("error", err))
			return err
		}

		_, err := dynamoClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName: aws.String(config.AgencyTableName),
			Key: map[string]types.AttributeValue{
				"pk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("invite#%s", message.Email),
				},
				"sk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("agency#%s", message.AgencyID),
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
		})

		if err != nil {
			logger.ErrorContext(ctx, "failed to update invite", slog.Any("error", err))
			return err
		}

		return nil
	}
}
