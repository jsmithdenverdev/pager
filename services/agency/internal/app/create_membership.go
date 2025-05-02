package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// CreateMembership creates a new membership in the agency.
func createMembership(config Config, logger *slog.Logger, dynamoClient *dynamodb.Client, snsClient *sns.Client) func(context.Context, events.SNSEntity) error {
	return func(ctx context.Context, record events.SNSEntity) error {
		var message struct {
			Email    string `json:"email"`
			AgencyID string `json:"agencyId"`
			UserID   string `json:"userId"`
		}

		if err := json.Unmarshal([]byte(record.Message), &message); err != nil {
			logger.ErrorContext(ctx, "failed to unmarshal message", slog.Any("error", err))
			return err
		}

		queryInviteResult, err := dynamoClient.Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(config.AgencyTableName),
			KeyConditionExpression: aws.String("pk = :pk AND sk = :sk"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("email#%s", message.Email),
				},
				":sk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("agency#%s", message.AgencyID),
				},
			},
		})

		if err != nil {
			logger.ErrorContext(ctx, "failed to query invite", slog.Any("error", err))
			return err
		}

		if len(queryInviteResult.Items) == 0 {
			logger.ErrorContext(ctx, "invite doesn't exist", slog.Any("messageId", record.MessageID))
			return nil
		}

		var invite invitation

		if err := attributevalue.UnmarshalMap(queryInviteResult.Items[0], &invite); err != nil {
			logger.ErrorContext(ctx, "failed to unmarshal invite", slog.Any("error", err))
			return err
		}

		if invite.Status != invitationStatusPending {
			logger.ErrorContext(ctx, "invite is not pending", slog.Any("messageId", record.MessageID))
			return nil
		}

		membershipAV, err := attributevalue.MarshalMap(membership{
			PK:         fmt.Sprintf("user#%s", message.UserID),
			SK:         fmt.Sprintf("agency#%s", message.AgencyID),
			Type:       entityTypeMembership,
			Role:       invite.Role,
			Status:     membershipStatusActive,
			Created:    time.Now(),
			Modified:   time.Now(),
			CreatedBy:  invite.CreatedBy,
			ModifiedBy: invite.ModifiedBy,
		})

		if err != nil {
			logger.ErrorContext(ctx, "failed to marshal membership", slog.Any("error", err))
			return err
		}

		_, err = dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(config.AgencyTableName),
			Item:      membershipAV,
		})

		if err != nil {
			logger.ErrorContext(ctx, "failed to put membership", slog.Any("error", err))
			return err
		}

		_, err = dynamoClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName: aws.String(config.AgencyTableName),
			Key: map[string]types.AttributeValue{
				"pk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("email#%s", message.Email),
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
					Value: "COMPLETED",
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
