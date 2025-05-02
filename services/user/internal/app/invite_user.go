package app

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
)

func inviteUser(config Config, logger *slog.Logger, dynamoClient *dynamodb.Client, snsClient *sns.Client) func(ctx context.Context, record events.SNSEntity) error {
	return func(ctx context.Context, record events.SNSEntity) error {
		var message struct {
			Email    string `json:"email"`
			AgencyID string `json:"agencyId"`
		}

		if err := json.Unmarshal([]byte(record.Message), &message); err != nil {
			logger.ErrorContext(ctx, "failed to unmarshal message", slog.Any("error", err))
			return err
		}

		logger.DebugContext(ctx, "looking up user", slog.Any("record", record), slog.Any("message", message))

		emailLookupResult, err := dynamoClient.Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(config.UserTableName),
			KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :skprefix)"),
			ExpressionAttributeValues: map[string]dynamodbtypes.AttributeValue{
				":pk": &dynamodbtypes.AttributeValueMemberS{
					Value: fmt.Sprintf("email#%s", message.Email),
				},
				":skprefix": &dynamodbtypes.AttributeValueMemberS{
					Value: "user#",
				},
			},
		})

		if err != nil {
			logger.ErrorContext(ctx, "failed to query user", slog.Any("error", err))
			return err
		}

		if len(emailLookupResult.Items) == 0 {
			logger.ErrorContext(ctx, "user doesn't exist (auth0 invite not implemented)", slog.Any("messageId", record.MessageID))
			return nil
		}

		var emailLookup lookup

		if err := attributevalue.UnmarshalMap(emailLookupResult.Items[0], &emailLookup); err != nil {
			logger.ErrorContext(ctx, "failed to unmarshal email lookup", slog.Any("error", err))
			return err
		}

		logger.DebugContext(ctx, "user invited")

		_, err = snsClient.Publish(ctx, &sns.PublishInput{
			TopicArn: aws.String(config.EventsTopicARN),
			Message:  aws.String(fmt.Sprintf(`{"email": "%s", "agencyId": "%s", "userId": "%s"}`, message.Email, message.AgencyID, emailLookup.UserID())),
			MessageAttributes: map[string]snstypes.MessageAttributeValue{
				"type": {
					DataType:    aws.String("String"),
					StringValue: aws.String("agency.membership.create"),
				},
			},
		})

		logger.DebugContext(ctx, "published event", slog.String("type", "agency.membership.create"))

		if err != nil {
			logger.ErrorContext(ctx, "failed to publish event", slog.Any("error", err))
			return err
		}

		return nil
	}
}
