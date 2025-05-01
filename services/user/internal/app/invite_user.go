package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
)

func inviteUser(config Config, logger *slog.Logger, dynamoClient *dynamodb.Client, snsClient *sns.Client) func(ctx context.Context, record events.SQSMessage) error {
	return func(ctx context.Context, record events.SQSMessage) error {
		var message struct {
			Email    string `json:"email"`
			AgencyID string `json:"agencyId"`
		}
		if err := json.Unmarshal([]byte(record.Body), &message); err != nil {
			logger.ErrorContext(ctx, "failed to unmarshal message", slog.Any("error", err))
			return err
		}
		userLookupResult, err := dynamoClient.Query(ctx, &dynamodb.QueryInput{
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
		if len(userLookupResult.Items) == 0 {
			logger.ErrorContext(ctx, "user doesn't exist (auth0 invites not implemented)", slog.Any("email", message.Email))
			return nil
		}

		var userLookup userLookup

		if err := attributevalue.UnmarshalMap(userLookupResult.Items[0], &userLookup); err != nil {
			logger.ErrorContext(ctx, "failed to unmarshal user lookup", slog.Any("error", err))
			return err
		}

		snsClient.Publish(ctx, &sns.PublishInput{
			TopicArn: aws.String(config.EventsTopicARN),
			Message:  aws.String(fmt.Sprintf(`{"email": "%s", "agencyId": "%s", "userId": "%s"}`, message.Email, message.AgencyID, strings.Split(userLookup.SK, "#")[1])),
			MessageAttributes: map[string]snstypes.MessageAttributeValue{
				"type": {
					DataType:    aws.String("String"),
					StringValue: aws.String("agency.membership.create"),
				},
			},
		})
		return nil
	}
}
