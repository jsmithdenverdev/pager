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
	"github.com/jsmithdenverdev/pager/pkg/identity"
)

func ensureUser(config Config, logger *slog.Logger, dynamoClient *dynamodb.Client, snsClient *sns.Client) func(context.Context, events.SNSEntity, int) error {
	return func(ctx context.Context, record events.SNSEntity, retryCount int) error {
		var message struct {
			Email    string        `json:"email"`
			Role     identity.Role `json:"role"`
			AgencyID string        `json:"agencyId"`
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
							StringValue: aws.String("user.user.ensure_and_invite_failed"),
						},
					},
				}); err != nil {
					logger.ErrorContext(ctx, "failed to publish ensure failed event", slog.Any("error", err))
					return err
				}
			}
			return err
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
			logger.ErrorContext(ctx, "failed to query user", slog.Any("error", err))
			if retryCount >= config.EventRetryCount {
				if _, err := snsClient.Publish(ctx, &sns.PublishInput{
					TopicArn: aws.String(config.EventsTopicARN),
					Message:  aws.String(fmt.Sprintf(`{"email": "%s", "agencyId": "%s"}`, message.Email, message.AgencyID)),
					MessageAttributes: map[string]snstypes.MessageAttributeValue{
						"type": {
							DataType:    aws.String("String"),
							StringValue: aws.String("user.user.ensure_and_invite_failed"),
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
		if len(lookupResult.Items) == 0 {
			logger.ErrorContext(ctx, "user doesn't exist (auth0 invite not implemented)", slog.Any("messageId", record.MessageID))
			if retryCount >= config.EventRetryCount {
				if _, err := snsClient.Publish(ctx, &sns.PublishInput{
					TopicArn: aws.String(config.EventsTopicARN),
					Message:  aws.String(fmt.Sprintf(`{"email": "%s", "agencyId": "%s"}`, message.Email, message.AgencyID)),
					MessageAttributes: map[string]snstypes.MessageAttributeValue{
						"type": {
							DataType:    aws.String("String"),
							StringValue: aws.String("user.user.ensure_and_invite_failed"),
						},
					},
				}); err != nil {
					logger.ErrorContext(ctx, "failed to publish ensure failed event", slog.Any("error", err))
					return err
				}
			}
			return fmt.Errorf("user doesn't exist (auth0 invite not implemented)")
		}

		var lookup lookup

		if err := attributevalue.UnmarshalMap(lookupResult.Items[0], &lookup); err != nil {
			logger.ErrorContext(ctx, "failed to unmarshal email lookup", slog.Any("error", err))
			if retryCount >= config.EventRetryCount {
				if _, err := snsClient.Publish(ctx, &sns.PublishInput{
					TopicArn: aws.String(config.EventsTopicARN),
					Message:  aws.String(fmt.Sprintf(`{"email": "%s", "agencyId": "%s"}`, message.Email, message.AgencyID)),
					MessageAttributes: map[string]snstypes.MessageAttributeValue{
						"type": {
							DataType:    aws.String("String"),
							StringValue: aws.String("user.user.ensure_and_invite_failed"),
						},
					},
				}); err != nil {
					logger.ErrorContext(ctx, "failed to publish ensure failed event", slog.Any("error", err))
					return err
				}
			}
			return err
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
			logger.ErrorContext(ctx, "failed to query user", slog.Any("error", err))
			if retryCount >= config.EventRetryCount {
				if _, err := snsClient.Publish(ctx, &sns.PublishInput{
					TopicArn: aws.String(config.EventsTopicARN),
					Message:  aws.String(fmt.Sprintf(`{"email": "%s", "agencyId": "%s"}`, message.Email, message.AgencyID)),
					MessageAttributes: map[string]snstypes.MessageAttributeValue{
						"type": {
							DataType:    aws.String("String"),
							StringValue: aws.String("user.user.ensure_and_invite_failed"),
						},
					},
				}); err != nil {
					logger.ErrorContext(ctx, "failed to publish ensure failed event", slog.Any("error", err))
					return err
				}
			}
			return err
		}

		if len(lookupResult.Items) == 0 {
			logger.ErrorContext(ctx, "user doesn't exist", slog.Any("messageId", record.MessageID))
			if retryCount >= config.EventRetryCount {
				if _, err := snsClient.Publish(ctx, &sns.PublishInput{
					TopicArn: aws.String(config.EventsTopicARN),
					Message:  aws.String(fmt.Sprintf(`{"email": "%s", "agencyId": "%s"}`, message.Email, message.AgencyID)),
					MessageAttributes: map[string]snstypes.MessageAttributeValue{
						"type": {
							DataType:    aws.String("String"),
							StringValue: aws.String("user.user.ensure_and_invite_failed"),
						},
					},
				}); err != nil {
					logger.ErrorContext(ctx, "failed to publish ensure failed event", slog.Any("error", err))
					return err
				}
			}
			return fmt.Errorf("user doesn't exist (auth0 invite not implemented)")
		}

		var user user

		if err := attributevalue.UnmarshalMap(userResult.Items[0], &user); err != nil {
			logger.ErrorContext(ctx, "failed to unmarshal user", slog.Any("error", err))
			if retryCount >= config.EventRetryCount {
				if _, err := snsClient.Publish(ctx, &sns.PublishInput{
					TopicArn: aws.String(config.EventsTopicARN),
					Message:  aws.String(fmt.Sprintf(`{"email": "%s", "agencyId": "%s"}`, message.Email, message.AgencyID)),
					MessageAttributes: map[string]snstypes.MessageAttributeValue{
						"type": {
							DataType:    aws.String("String"),
							StringValue: aws.String("user.user.ensure_and_invite_failed"),
						},
					},
				}); err != nil {
					logger.ErrorContext(ctx, "failed to publish ensure failed event", slog.Any("error", err))
					return err
				}
			}
			return err
		}

		user.Memberships[message.AgencyID] = message.Role

		memberships, err := attributevalue.MarshalMap(user.Memberships)
		if err != nil {
			logger.ErrorContext(ctx, "failed to marshal memberships", slog.Any("error", err))
			if retryCount >= config.EventRetryCount {
				if _, err := snsClient.Publish(ctx, &sns.PublishInput{
					TopicArn: aws.String(config.EventsTopicARN),
					Message:  aws.String(fmt.Sprintf(`{"email": "%s", "agencyId": "%s"}`, message.Email, message.AgencyID)),
					MessageAttributes: map[string]snstypes.MessageAttributeValue{
						"type": {
							DataType:    aws.String("String"),
							StringValue: aws.String("user.user.ensure_and_invite_failed"),
						},
					},
				}); err != nil {
					logger.ErrorContext(ctx, "failed to publish ensure failed event", slog.Any("error", err))
					return err
				}
			}
			return err
		}

		_, err = dynamoClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName: aws.String(config.UserTableName),
			Key: map[string]dynamodbtypes.AttributeValue{
				"pk": &dynamodbtypes.AttributeValueMemberS{
					Value: fmt.Sprintf("user#%s", lookup.UserID),
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
			logger.ErrorContext(ctx, "failed to update user", slog.Any("error", err))
			if retryCount >= config.EventRetryCount {
				if _, err := snsClient.Publish(ctx, &sns.PublishInput{
					TopicArn: aws.String(config.EventsTopicARN),
					Message:  aws.String(fmt.Sprintf(`{"email": "%s", "agencyId": "%s"}`, message.Email, message.AgencyID)),
					MessageAttributes: map[string]snstypes.MessageAttributeValue{
						"type": {
							DataType:    aws.String("String"),
							StringValue: aws.String("user.user.ensure_and_invite_failed"),
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
			Message:  aws.String(fmt.Sprintf(`{"email": "%s", "agencyId": "%s", "userId": "%s"}`, message.Email, message.AgencyID, lookup.UserID)),
			MessageAttributes: map[string]snstypes.MessageAttributeValue{
				"type": {
					DataType:    aws.String("String"),
					StringValue: aws.String("user.user.ensure_and_invited"),
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
							StringValue: aws.String("user.user.ensure_and_invite_failed"),
						},
					},
				}); err != nil {
					logger.ErrorContext(ctx, "failed to publish ensure failed event", slog.Any("error", err))
					return err
				}
			}
			return err
		}

		logger.DebugContext(ctx, "published event", slog.String("type", "user.user.ensure_and_invited"))

		return nil
	}
}
