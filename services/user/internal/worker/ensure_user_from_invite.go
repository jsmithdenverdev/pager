package worker

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/auth0/go-auth0/authentication"
	"github.com/auth0/go-auth0/authentication/database"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/google/uuid"
	"github.com/jsmithdenverdev/pager/pkg/identity"
	"github.com/jsmithdenverdev/pager/services/user/internal/models"
)

func ensureUserFromInvite(config Config, logger *slog.Logger, dynamoClient *dynamodb.Client, snsClient *sns.Client, auth0Client *authentication.Authentication) func(context.Context, events.SNSEntity, int) error {
	type message struct {
		Email    string `json:"email"`
		AgencyID string `json:"agencyId"`
	}

	logAndHandleError := eventProcessorErrorHandler(config, logger, snsClient, evtEnsureInviteFailed)

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

		// User doesn't exist, create them in auth0
		if len(lookupResult.Items) == 0 {
			// Generate a secure temporary password using a GUID and hashing
			tempGUID := uuid.New().String()
			hash := sha256.Sum256([]byte(tempGUID))
			tempPassword := hex.EncodeToString(hash[:])

			// Create the user in auth0
			createUserResponse, err := auth0Client.Database.Signup(ctx, database.SignupRequest{
				Connection: config.Auth0Connection,
				Email:      message.Email,
				Password:   tempPassword,
			})
			if err != nil {
				return logAndHandleError(ctx, retryCount, "failed to ensure user from invite", message, err)
			}

			// Force a password reset for the user
			_, err = auth0Client.Database.ChangePassword(ctx, database.ChangePasswordRequest{
				Connection: config.Auth0Connection,
				Email:      message.Email,
			})
			if err != nil {
				return logAndHandleError(ctx, retryCount, "failed to ensure user from invite", message, err)
			}

			now := time.Now()

			lookup := models.Lookup{
				Keys: models.Keys{
					PK: fmt.Sprintf("lookup#%s", message.Email),
					SK: "lookup",
				},
				Auditable: models.Auditable{
					Created:    now,
					Modified:   now,
					CreatedBy:  "system",
					ModifiedBy: "system",
				},
				UserID: createUserResponse.ID,
			}

			lookupAV, err := attributevalue.MarshalMap(lookup)
			if err != nil {
				return logAndHandleError(ctx, retryCount, "failed to ensure user from invite", message, err)
			}

			user := models.User{
				Keys: models.Keys{
					PK: fmt.Sprintf("user#%s", createUserResponse.ID),
					SK: "meta",
				},
				Auditable: models.Auditable{
					Created:    now,
					Modified:   now,
					CreatedBy:  "system",
					ModifiedBy: "system",
				},
				Name:        message.Email,
				Email:       message.Email,
				Memberships: map[string]identity.Role{},
			}

			userAV, err := attributevalue.MarshalMap(user)
			if err != nil {
				return logAndHandleError(ctx, retryCount, "failed to ensure user from invite", message, err)
			}

			if _, err := dynamoClient.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
				TransactItems: []dynamodbtypes.TransactWriteItem{
					{
						Put: &dynamodbtypes.Put{
							TableName: aws.String(config.UserTableName),
							Item:      lookupAV,
						},
					},
					{
						Put: &dynamodbtypes.Put{
							TableName: aws.String(config.UserTableName),
							Item:      userAV,
						},
					},
				},
			}); err != nil {
				return logAndHandleError(ctx, retryCount, "failed to ensure user from invite", message, err)
			}

			if _, err := snsClient.Publish(ctx, &sns.PublishInput{
				TopicArn: aws.String(config.EventsTopicARN),
				Message:  aws.String(fmt.Sprintf(`{"email": "%s", "agencyId": "%s", "userId": "%s"}`, message.Email, message.AgencyID, lookup.UserID)),
				MessageAttributes: map[string]snstypes.MessageAttributeValue{
					"type": {
						DataType:    aws.String("String"),
						StringValue: aws.String(evtInviteTargetEnsured),
					},
				},
			}); err != nil {
				return logAndHandleError(ctx, retryCount, "failed to ensure user from invite", message, err)
			}

			logger.DebugContext(ctx, "published event", slog.String("type", evtInviteTargetEnsured))

			return nil
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

		if _, err := snsClient.Publish(ctx, &sns.PublishInput{
			TopicArn: aws.String(config.EventsTopicARN),
			Message:  aws.String(fmt.Sprintf(`{"email": "%s", "agencyId": "%s", "userId": "%s"}`, message.Email, message.AgencyID, lookup.UserID)),
			MessageAttributes: map[string]snstypes.MessageAttributeValue{
				"type": {
					DataType:    aws.String("String"),
					StringValue: aws.String(evtInviteTargetEnsured),
				},
			},
		}); err != nil {
			return logAndHandleError(ctx, retryCount, "failed to ensure user from invite", message, err)
		}

		logger.DebugContext(ctx, "published event", slog.String("type", evtInviteTargetEnsured))

		return nil
	}
}
