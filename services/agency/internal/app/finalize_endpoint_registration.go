package app

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
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

func finalizeEndpointRegistration(config Config, logger *slog.Logger, dynamoClient *dynamodb.Client, snsClient *sns.Client) func(context.Context, events.SNSEntity) error {
	return func(ctx context.Context, record events.SNSEntity) error {
		var message struct {
			AgencyID         string `json:"agencyId"`
			RegistrationCode string `json:"registrationCode"`
			EndpointID       string `json:"endpointId"`
		}

		if err := json.Unmarshal([]byte(record.Message), &message); err != nil {
			logger.ErrorContext(ctx, "failed to unmarshal message", slog.Any("error", err))
			return err
		}

		queryRegistrationResult, err := dynamoClient.Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(config.AgencyTableName),
			KeyConditionExpression: aws.String("pk = :pk AND sk = :sk"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("agency#%s", message.AgencyID),
				},
				":sk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("registration#%s", message.RegistrationCode),
				},
			},
		})

		if err != nil {
			logger.ErrorContext(ctx, "failed to query registration", slog.Any("error", err))
			return err
		}

		if len(queryRegistrationResult.Items) == 0 {
			logger.ErrorContext(ctx, "registration doesn't exist", slog.Any("registrationCode", message.RegistrationCode))
			return errors.New("registration doesn't exist")
		}

		var registration endpointRegistration

		if err := attributevalue.UnmarshalMap(queryRegistrationResult.Items[0], &registration); err != nil {
			logger.ErrorContext(ctx, "failed to unmarshal registration", slog.Any("error", err))
			return err
		}

		if registration.Status != registrationStatusPending {
			logger.ErrorContext(ctx, "registration is not pending", slog.Any("registrationCode", message.RegistrationCode), slog.String("status", registration.Status))
			return errors.New("registration is not pending")
		}

		_, err = dynamoClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName: aws.String(config.AgencyTableName),
			Key: map[string]types.AttributeValue{
				"pk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("agency#%s", message.AgencyID),
				},
				"sk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("registration#%s", message.RegistrationCode),
				},
			},
			UpdateExpression: aws.String("set #status = :status, #endpointId = :endpointId"),
			ExpressionAttributeNames: map[string]string{
				"#status":     "status",
				"#endpointId": "endpointId",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":status": &types.AttributeValueMemberS{
					Value: "COMPLETED",
				},
				":endpointId": &types.AttributeValueMemberS{
					Value: message.EndpointID,
				},
			},
		})

		if err != nil {
			logger.ErrorContext(ctx, "failed to update registration", slog.Any("error", err))
			return err
		}

		return nil
	}
}
