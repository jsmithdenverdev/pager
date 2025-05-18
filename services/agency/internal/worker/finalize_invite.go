package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/jsmithdenverdev/pager/services/agency/internal/models"
)

// finalizeInvite creates a new membership in the agency related to
// an originating invitation.
func finalizeInvite(config Config, logger *slog.Logger, dynamoClient *dynamodb.Client, snsClient *sns.Client) func(context.Context, events.SNSEntity, int) error {
	type message struct {
		Email    string `json:"email"`
		AgencyID string `json:"agencyId"`
		UserID   string `json:"userId"`
	}

	logAndHandleError := eventProcessorErrorHandler(config, logger, snsClient, evtMembershipCreateFailed)

	return func(ctx context.Context, record events.SNSEntity, retryCount int) error {
		var message message

		if err := json.Unmarshal([]byte(record.Message), &message); err != nil {
			return logAndHandleError(ctx, retryCount, "failed to create membership", message, err)
		}

		queryInviteResult, err := dynamoClient.Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(config.AgencyTableName),
			KeyConditionExpression: aws.String("#pk = :pk AND #sk = :sk"),
			ExpressionAttributeNames: map[string]string{
				"#pk": "pk",
				"#sk": "sk",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("invite#%s", message.Email),
				},
				":sk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("agency#%s", message.AgencyID),
				},
			},
		})

		if err != nil {
			return logAndHandleError(ctx, retryCount, "failed to create membership", message, err)
		}

		if len(queryInviteResult.Items) == 0 {
			return logAndHandleError(ctx, retryCount, "failed to create membership", message, errors.New("invite doesn't exist"))
		}

		var invite models.Invitation

		if err := attributevalue.UnmarshalMap(queryInviteResult.Items[0], &invite); err != nil {
			return logAndHandleError(ctx, retryCount, "failed to create membership", message, err)
		}

		if invite.Status != models.InvitationStatusPending {
			return logAndHandleError(ctx, retryCount, "failed to create membership", message, errors.New("invite is not pending"))
		}

		membershipAV, err := attributevalue.MarshalMap(models.Membership{
			PK:         fmt.Sprintf("user#%s", message.UserID),
			SK:         fmt.Sprintf("agency#%s", message.AgencyID),
			Type:       models.EntityTypeMembership,
			Role:       invite.Role,
			Status:     models.MembershipStatusActive,
			Created:    time.Now(),
			Modified:   time.Now(),
			CreatedBy:  invite.CreatedBy,
			ModifiedBy: invite.ModifiedBy,
		})

		if err != nil {
			return logAndHandleError(ctx, retryCount, "failed to create membership", message, err)
		}

		membershipInverseAV, err := attributevalue.MarshalMap(models.Membership{
			PK:         fmt.Sprintf("agency#%s", message.AgencyID),
			SK:         fmt.Sprintf("user#%s", message.UserID),
			Type:       models.EntityTypeMembership,
			Role:       invite.Role,
			Status:     models.MembershipStatusActive,
			Created:    time.Now(),
			Modified:   time.Now(),
			CreatedBy:  invite.CreatedBy,
			ModifiedBy: invite.ModifiedBy,
		})

		if err != nil {
			return logAndHandleError(ctx, retryCount, "failed to create membership", message, err)
		}

		if _, err := dynamoClient.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
			TransactItems: []types.TransactWriteItem{
				{
					Put: &types.Put{
						TableName: aws.String(config.AgencyTableName),
						Item:      membershipAV,
					},
				},
				{
					Put: &types.Put{
						TableName: aws.String(config.AgencyTableName),
						Item:      membershipInverseAV,
					},
				},
				{
					Update: &types.Update{
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
								Value: "COMPLETED",
							},
						},
					},
				},
			},
		}); err != nil {
			return logAndHandleError(ctx, retryCount, "failed to create membership", message, err)
		}

		if _, err := snsClient.Publish(ctx, &sns.PublishInput{
			TopicArn: aws.String(config.EventsTopicARN),
			Message:  aws.String(fmt.Sprintf(`{"userId": "%s", "agencyId": "%s", "role": "%s"}`, message.UserID, message.AgencyID, invite.Role)),
			MessageAttributes: map[string]snstypes.MessageAttributeValue{
				"type": {
					DataType:    aws.String("String"),
					StringValue: aws.String(evtMembershipCreated),
				},
			},
		}); err != nil {
			return logAndHandleError(ctx, retryCount, "failed to publish create membership event", message, err)
		}

		logger.DebugContext(ctx, "published event", slog.String("type", evtMembershipCreated))

		return nil
	}
}
