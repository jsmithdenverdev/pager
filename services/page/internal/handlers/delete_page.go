package handlers

import (
	"context"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"
	"github.com/jsmithdenverdev/pager/pkg/apigateway"
	"github.com/jsmithdenverdev/pager/pkg/problemdetail"
)

func DeletePage(logger *slog.Logger) func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		var (
			errEncoder = apigateway.NewProblemDetailEncoder(apigateway.WithLogger[problemdetail.ProblemDetailer](logger))
		)

		return errEncoder.EncodeInternalServerError(ctx), nil
	}
}
