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
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/jsmithdenverdev/pager/services/endpoint/internal/models"
)

func ensureAndRegisterEndpoint(config Config, logger *slog.Logger, dynamoClient *dynamodb.Client, snsClient *sns.Client) func(ctx context.Context, snsRecord events.SNSEntity) error {
	return func(ctx context.Context, snsRecord events.SNSEntity) error {
		var message struct {
			AgencyID         string `json:"agencyId"`
			RegistrationCode string `json:"registrationCode"`
		}

		if err := json.Unmarshal([]byte(snsRecord.Message), &message); err != nil {
			logger.ErrorContext(ctx, "failed to unmarshal message", slog.Any("error", err))
			return err
		}

		queryRegistrationCodeResult, err := dynamoClient.Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(config.EndpointTableName),
			KeyConditionExpression: aws.String("#pk = :pk AND #sk = :sk"),
			ExpressionAttributeNames: map[string]string{
				"#pk": "pk",
				"#sk": "sk",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("rc#%s", message.RegistrationCode),
				},
				":sk": &types.AttributeValueMemberS{
					Value: "registrationcode",
				},
			},
		})

		if err != nil {
			logger.ErrorContext(ctx, "failed to query registration code", slog.Any("error", err))
			return err
		}

		if len(queryRegistrationCodeResult.Items) == 0 {
			logger.ErrorContext(ctx, "registration code doesn't exist", slog.Any("registrationCode", message.RegistrationCode))
			return fmt.Errorf("registration code doesn't exist")
		}

		var rc models.RegistrationCode

		if err := attributevalue.UnmarshalMap(queryRegistrationCodeResult.Items[0], &rc); err != nil {
			logger.ErrorContext(ctx, "failed to unmarshal registration code", slog.Any("error", err))
			return err
		}

		queryEndpointResult, err := dynamoClient.Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(config.EndpointTableName),
			KeyConditionExpression: aws.String("#pk = :pk AND #sk = :sk"),
			ExpressionAttributeNames: map[string]string{
				"#pk": "pk",
				"#sk": "sk",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("endpoint#%s", rc.EndpointID),
				},
				":sk": &types.AttributeValueMemberS{
					Value: "meta",
				},
			},
		})

		if err != nil {
			logger.ErrorContext(ctx, "failed to query endpoint", slog.Any("error", err))
			return err
		}

		if len(queryEndpointResult.Items) == 0 {
			logger.ErrorContext(ctx, "endpoint doesn't exist", slog.Any("userId", rc.UserID), slog.Any("endpointId", rc.EndpointID))
			return fmt.Errorf("endpoint doesn't exist")
		}

		var endpoint models.Endpoint

		if err := attributevalue.UnmarshalMap(queryEndpointResult.Items[0], &endpoint); err != nil {
			logger.ErrorContext(ctx, "failed to unmarshal endpoint", slog.Any("error", err))
			return err
		}

		endpoint.Registrations = append(endpoint.Registrations, message.AgencyID)

		_, err = dynamoClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName: aws.String(config.EndpointTableName),
			Key: map[string]types.AttributeValue{
				"pk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("endpoint#%s", rc.EndpointID),
				},
				"sk": &types.AttributeValueMemberS{
					Value: "meta",
				},
			},
			UpdateExpression: aws.String("SET #registrations = :registrations"),
			ExpressionAttributeNames: map[string]string{
				"#registrations": "registrations",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":registrations": &types.AttributeValueMemberSS{
					Value: endpoint.Registrations,
				},
			},
		})

		if err != nil {
			logger.ErrorContext(ctx, "failed to update endpoint", slog.Any("error", err))
			return err
		}

		messageBody, err := json.Marshal(struct {
			RegistrationCode string `json:"registrationCode"`
			AgencyID         string `json:"agencyId"`
			EndpointID       string `json:"endpointId"`
		}{
			RegistrationCode: message.RegistrationCode,
			AgencyID:         message.AgencyID,
			EndpointID:       rc.EndpointID,
		})

		if err != nil {
			logger.ErrorContext(ctx, "failed to marshal SNS message", slog.Any("error", err))
			return err
		}

		if _, err = snsClient.Publish(ctx, &sns.PublishInput{
			TopicArn: aws.String(config.EventsTopicARN),
			Message:  aws.String(string(messageBody)),
			MessageAttributes: map[string]snstypes.MessageAttributeValue{
				"type": {
					DataType:    aws.String("String"),
					StringValue: aws.String("endpoint.endpoint.ensured_and_registered"),
				},
			},
		}); err != nil {
			logger.ErrorContext(ctx, "failed to publish SNS message", slog.Any("error", err))
			return err
		}

		return nil
	}
}
