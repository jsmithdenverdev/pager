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
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/jsmithdenverdev/pager/services/user/internal/models"
)

func ensureUserFromInvite(config Config, logger *slog.Logger, dynamoClient *dynamodb.Client, snsClient *sns.Client) func(context.Context, events.SNSEntity, int) error {
	type message struct {
		Email    string `json:"email"`
		AgencyID string `json:"agencyId"`
	}

	logAndHandleError := eventProcessorErrorHandler(config, logger, snsClient, evtUserInviteEnsureFailed)

	return func(ctx context.Context, record events.SNSEntity, retryCount int) error {
		var message message

		if err := json.Unmarshal([]byte(record.Message), &message); err != nil {
			return logAndHandleError(ctx, retryCount, "failed to ensure user from invite", message, err)
		}

		lookupResult, err := dynamoClient.Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(config.UserTableName),
			KeyConditionExpression: aws.String("#pk = :pk AND #sk = :sk"),
			ExpressionAttributeNames: map[string]string{
				"#pk": "pk",
				"#sk": "sk",
			},
			ExpressionAttributeValues: map[string]dynamodbtypes.AttributeValue{
				":pk": &dynamodbtypes.AttributeValueMemberS{
					Value: fmt.Sprintf("lookup#%s", message.Email),
				},
				":sk": &dynamodbtypes.AttributeValueMemberS{
					Value: "lookup",
				},
			},
		})

		if err != nil {
			return logAndHandleError(ctx, retryCount, "failed to ensure user from invite", message, err)
		}

		// TODO: create a user in auth0 if they don't exist
		if len(lookupResult.Items) == 0 {
			return logAndHandleError(ctx, retryCount, "failed to ensure user from invite", message, errors.New("user doesn't exist (auth0 invite not implemented)"))
		}

		var lookup models.Lookup

		if err := attributevalue.UnmarshalMap(lookupResult.Items[0], &lookup); err != nil {
			return logAndHandleError(ctx, retryCount, "failed to ensure user from invite", message, err)
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
					Value: fmt.Sprintf("user#%s", lookup.UserID),
				},
				":sk": &dynamodbtypes.AttributeValueMemberS{
					Value: "meta",
				},
			},
		})

		if err != nil {
			return logAndHandleError(ctx, retryCount, "failed to ensure user from invite", message, err)
		}

		if len(userResult.Items) == 0 {
			return logAndHandleError(ctx, retryCount, "failed to ensure user from invite", message, errors.New("user doesn't exist (auth0 invite not implemented)"))
		}

		var user models.User

		if err := attributevalue.UnmarshalMap(userResult.Items[0], &user); err != nil {
			return logAndHandleError(ctx, retryCount, "failed to ensure user from invite", message, err)
		}

		_, err = snsClient.Publish(ctx, &sns.PublishInput{
			TopicArn: aws.String(config.EventsTopicARN),
			Message:  aws.String(fmt.Sprintf(`{"email": "%s", "agencyId": "%s", "userId": "%s"}`, message.Email, message.AgencyID, lookup.UserID)),
			MessageAttributes: map[string]snstypes.MessageAttributeValue{
				"type": {
					DataType:    aws.String("String"),
					StringValue: aws.String(evtUserInviteEnsured),
				},
			},
		})

		if err != nil {
			return logAndHandleError(ctx, retryCount, "failed to ensure user from invite", message, err)
		}

		logger.DebugContext(ctx, "published event", slog.String("type", evtUserInviteEnsured))

		return nil
	}
}
