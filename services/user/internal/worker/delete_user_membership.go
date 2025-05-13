package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/jsmithdenverdev/pager/services/user/internal/models"
)

func deleteUserMembership(config Config, logger *slog.Logger, dynamoClient *dynamodb.Client, snsClient *sns.Client) func(context.Context, events.SNSEntity, int) error {
	type message struct {
		UserID   string `json:"userId"`
		AgencyID string `json:"agencyId"`
	}

	logAndHandleError := eventProcessorErrorHandler(config, logger, snsClient, evtMembershipDeleteFailed)

	return func(ctx context.Context, record events.SNSEntity, retryCount int) error {
		var message message

		if err := json.Unmarshal([]byte(record.Message), &message); err != nil {
			// TODO
			// If we can't unmarshal the request, we won't have the metdata needed to
			// create a message to publish back as a failure.
			// E.g., this should not be publishing to the sns topic.
			return logAndHandleError(ctx, retryCount, "failed to delete user membership", message, err)
		}

		userResult, err := dynamoClient.Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(config.UserTableName),
			KeyConditionExpression: aws.String("#pk = :pk AND #sk = :sk"),
			ExpressionAttributeNames: map[string]string{
				"#pk": "pk",
				"#sk": "sk",
			},
			ExpressionAttributeValues: map[string]dynamodbtypes.AttributeValue{
				":pk": &dynamodbtypes.AttributeValueMemberS{
					Value: fmt.Sprintf("user#%s", message.UserID),
				},
				":sk": &dynamodbtypes.AttributeValueMemberS{
					Value: "meta",
				},
			},
		})

		if err != nil {
			return logAndHandleError(ctx, retryCount, "failed to delete user membership", message, err)
		}

		if len(userResult.Items) == 0 {
			return logAndHandleError(ctx, retryCount, "failed to delete user membership", message, err)
		}

		var user models.User

		if err := attributevalue.UnmarshalMap(userResult.Items[0], &user); err != nil {
			return logAndHandleError(ctx, retryCount, "failed to delete user membership", message, err)
		}

		delete(user.Memberships, message.AgencyID)

		memberships, err := attributevalue.MarshalMap(user.Memberships)
		if err != nil {
			return logAndHandleError(ctx, retryCount, "failed to delete user membership", message, err)
		}

		_, err = dynamoClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName: aws.String(config.UserTableName),
			Key: map[string]dynamodbtypes.AttributeValue{
				"pk": &dynamodbtypes.AttributeValueMemberS{
					Value: fmt.Sprintf("user#%s", message.UserID),
				},
				"sk": &dynamodbtypes.AttributeValueMemberS{
					Value: "meta",
				},
			},
			UpdateExpression: aws.String("set #memberships = :memberships"),
			ExpressionAttributeNames: map[string]string{
				"#memberships": "memberships",
			},
			ExpressionAttributeValues: map[string]dynamodbtypes.AttributeValue{
				":memberships": &dynamodbtypes.AttributeValueMemberM{
					Value: memberships,
				},
			},
		})

		if err != nil {
			return logAndHandleError(ctx, retryCount, "failed to delete user membership", message, err)
		}

		_, err = snsClient.Publish(ctx, &sns.PublishInput{
			TopicArn: aws.String(config.EventsTopicARN),
			Message:  aws.String(fmt.Sprintf(`{"userId": "%s", "agencyId": "%s"}`, message.UserID, message.AgencyID)),
			MessageAttributes: map[string]snstypes.MessageAttributeValue{
				"type": {
					DataType:    aws.String("String"),
					StringValue: aws.String(evtMembershipDeleted),
				},
			},
		})

		if err != nil {
			return logAndHandleError(ctx, retryCount, "failed to delete user membership", message, err)
		}

		logger.DebugContext(ctx, "published event", slog.String("type", evtMembershipDeleted))

		return nil
	}
}
