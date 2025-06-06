package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/jsmithdenverdev/pager/services/endpoint/internal/models"
)

func deliverToEndpoints(config Config, logger *slog.Logger, dynamoClient *dynamodb.Client, snsClient *sns.Client) func(ctx context.Context, snsRecord events.SNSEntity, retryCount int) error {
	type message struct {
		AgencyID string `json:"agencyId"`
		Title    string `json:"title"`
		PageID   string `json:"pageId"`
	}

	logAndHandleError := eventProcessorErrorHandler(config, logger, snsClient, evtEndpointResolutionFailed)

	return func(ctx context.Context, snsRecord events.SNSEntity, retryCount int) error {
		var message message

		if err := json.Unmarshal([]byte(snsRecord.Message), &message); err != nil {
			return logAndHandleError(ctx, retryCount, "failed to unmarshal endpoint.deliver message", message, err)
		}

		queryEndpointsResult, err := dynamoClient.Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(config.EndpointTableName),
			KeyConditionExpression: aws.String("#pk = :pk AND begins_with(#sk, :skprefix)"),
			ExpressionAttributeNames: map[string]string{
				"#pk": "pk",
				"#sk": "sk",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("agency#%s", message.AgencyID),
				},
				":skprefix": &types.AttributeValueMemberS{
					Value: "endpoint#",
				},
			},
		})

		if err != nil {
			return logAndHandleError(ctx, retryCount, "failed to query endpoints", message, err)
		}

		var registeredEndpoints []models.Registration

		if err := attributevalue.UnmarshalListOfMaps(queryEndpointsResult.Items, &registeredEndpoints); err != nil {
			return logAndHandleError(ctx, retryCount, "failed to unmarshal endpoints for agency", message, err)
		}

		logger.InfoContext(
			ctx,
			"delivering to endpoints",
			slog.String("pageId", message.PageID),
			slog.String("agencyId", message.AgencyID),
			slog.String("title", message.Title),
			slog.Any("endpoints", registeredEndpoints))

		for _, registeredEndpoint := range registeredEndpoints {
			// TEST CODE
			msg, err := json.Marshal(struct {
				Title      string `json:"title"`
				PageID     string `json:"pageId"`
				EndpointID string `json:"endpointId"`
			}{
				message.Title,
				message.PageID,
				registeredEndpoint.PK,
			})
			if err != nil {
				return logAndHandleError(ctx, retryCount, "failed to deliver to endpoint", message, err)
			}
			if registeredEndpoint.EndpointType == models.EndpointTypeWebhook {
				http.Post(registeredEndpoint.URL, "application/json", bytes.NewBuffer(msg))
			}
		}

		return nil
	}
}
