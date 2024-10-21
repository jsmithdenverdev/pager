package middleware

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
)

type (
	APIGatewayLambdaHandler    = func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)
	APIGatewayLambdaMiddleware = func(next APIGatewayLambdaHandler) APIGatewayLambdaHandler
)
