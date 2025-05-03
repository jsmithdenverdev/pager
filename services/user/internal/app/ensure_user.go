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

func ensureUser(config Config, logger *slog.Logger, dynamoClient *dynamodb.Client, snsClient *sns.Client) func(context.Context, events.SNSEntity, int) error {
	return func(ctx context.Context, record events.SNSEntity, retryCount int) error {
		var message struct {
			Email    string `json:"email"`
			AgencyID string `json:"agencyId"`
		}

		if err := json.Unmarshal([]byte(record.Message), &message); err != nil {
			logger.ErrorContext(ctx, "failed to unmarshal message", slog.Any("error", err))
			if retryCount >= config.EventRetryCount {
				if _, err := snsClient.Publish(ctx, &sns.PublishInput{
					TopicArn: aws.String(config.EventsTopicARN),
					Message:  aws.String(fmt.Sprintf(`{"email": "%s", "agencyId": "%s"}`, message.Email, message.AgencyID)),
					MessageAttributes: map[string]snstypes.MessageAttributeValue{
						"type": {
							DataType:    aws.String("String"),
							StringValue: aws.String("user.user.ensure_failed"),
						},
					},
				}); err != nil {
					logger.ErrorContext(ctx, "failed to publish ensure failed event", slog.Any("error", err))
					return err
				}
			}
			return err
		}

		ensureResult, err := dynamoClient.Query(ctx, &dynamodb.QueryInput{
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
			if retryCount >= config.EventRetryCount {
				if _, err := snsClient.Publish(ctx, &sns.PublishInput{
					TopicArn: aws.String(config.EventsTopicARN),
					Message:  aws.String(fmt.Sprintf(`{"email": "%s", "agencyId": "%s"}`, message.Email, message.AgencyID)),
					MessageAttributes: map[string]snstypes.MessageAttributeValue{
						"type": {
							DataType:    aws.String("String"),
							StringValue: aws.String("user.user.ensure_failed"),
						},
					},
				}); err != nil {
					logger.ErrorContext(ctx, "failed to publish ensure failed event", slog.Any("error", err))
					return err
				}
			}
			return err
		}

		// TODO: create a user in auth0 if they don't exist
		if len(ensureResult.Items) == 0 {
			logger.ErrorContext(ctx, "user doesn't exist (auth0 invite not implemented)", slog.Any("messageId", record.MessageID))
			if retryCount >= config.EventRetryCount {
				if _, err := snsClient.Publish(ctx, &sns.PublishInput{
					TopicArn: aws.String(config.EventsTopicARN),
					Message:  aws.String(fmt.Sprintf(`{"email": "%s", "agencyId": "%s"}`, message.Email, message.AgencyID)),
					MessageAttributes: map[string]snstypes.MessageAttributeValue{
						"type": {
							DataType:    aws.String("String"),
							StringValue: aws.String("user.user.ensure_failed"),
						},
					},
				}); err != nil {
					logger.ErrorContext(ctx, "failed to publish ensure failed event", slog.Any("error", err))
					return err
				}
			}
			return fmt.Errorf("user doesn't exist (auth0 invite not implemented)")
		}

		var emailLookup lookup

		if err := attributevalue.UnmarshalMap(ensureResult.Items[0], &emailLookup); err != nil {
			logger.ErrorContext(ctx, "failed to unmarshal email lookup", slog.Any("error", err))
			if retryCount >= config.EventRetryCount {
				if _, err := snsClient.Publish(ctx, &sns.PublishInput{
					TopicArn: aws.String(config.EventsTopicARN),
					Message:  aws.String(fmt.Sprintf(`{"email": "%s", "agencyId": "%s"}`, message.Email, message.AgencyID)),
					MessageAttributes: map[string]snstypes.MessageAttributeValue{
						"type": {
							DataType:    aws.String("String"),
							StringValue: aws.String("user.user.ensure_failed"),
						},
					},
				}); err != nil {
					logger.ErrorContext(ctx, "failed to publish ensure failed event", slog.Any("error", err))
					return err
				}
			}
			return err
		}

		_, err = snsClient.Publish(ctx, &sns.PublishInput{
			TopicArn: aws.String(config.EventsTopicARN),
			Message:  aws.String(fmt.Sprintf(`{"email": "%s", "agencyId": "%s", "userId": "%s"}`, message.Email, message.AgencyID, emailLookup.UserID())),
			MessageAttributes: map[string]snstypes.MessageAttributeValue{
				"type": {
					DataType:    aws.String("String"),
					StringValue: aws.String("user.user.ensured"),
				},
			},
		})

		if err != nil {
			logger.ErrorContext(ctx, "failed to publish event", slog.Any("error", err))
			if retryCount >= config.EventRetryCount {
				if _, err := snsClient.Publish(ctx, &sns.PublishInput{
					TopicArn: aws.String(config.EventsTopicARN),
					Message:  aws.String(fmt.Sprintf(`{"email": "%s", "agencyId": "%s"}`, message.Email, message.AgencyID)),
					MessageAttributes: map[string]snstypes.MessageAttributeValue{
						"type": {
							DataType:    aws.String("String"),
							StringValue: aws.String("user.user.ensure_failed"),
						},
					},
				}); err != nil {
					logger.ErrorContext(ctx, "failed to publish ensure failed event", slog.Any("error", err))
					return err
				}
			}
			return err
		}

		logger.DebugContext(ctx, "published event", slog.String("type", "user.user.ensured"))

		return nil
	}
}
