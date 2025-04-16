package app

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func UserInfo(config Config, logger *slog.Logger, client *dynamodb.Client) func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	type userInfoResponse struct {
		Email  string `json:"email"`
		IDPID  string `json:"idpId"`
		Status string `json:"status"`
	}

	return func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		logger.InfoContext(ctx, "user-info", slog.Any("headers", event.Headers))
		userId, ok := event.Headers["x-pager-userid"]
		if !ok {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusOK,
				Body:       "",
			}, nil
		}

		row, err := client.GetItem(ctx, &dynamodb.GetItemInput{
			TableName: aws.String(config.TableName),
			Key: map[string]types.AttributeValue{
				"idpid": &types.AttributeValueMemberS{
					Value: userId,
				},
			},
		})

		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       "internal server error",
			}, err
		}

		var user user

		if err := attributevalue.UnmarshalMap(row.Item, &user); err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       "internal server error",
			}, err
		}

		resp, err := json.Marshal(userInfoResponse{
			Email:  user.Email,
			IDPID:  user.IDPID,
			Status: user.Status,
		})

		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       "internal server error",
			}, err
		}

		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       string(resp),
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}, nil
	}
}
