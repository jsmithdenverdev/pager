package alert

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

func HandleAlert() func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       "Alert",
		}, nil
	}
}
