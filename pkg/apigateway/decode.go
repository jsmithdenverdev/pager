package apigateway

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/jsmithdenverdev/pager/pkg/valid"
)

func DecodeValid[TOut valid.Validator](ctx context.Context, in events.APIGatewayProxyRequest) (TOut, error) {
	var out TOut
	if err := json.Unmarshal([]byte(in.Body), &out); err != nil {
		return out, err
	}

	if problems := out.Valid(ctx); len(problems) > 0 {
		return *new(TOut), valid.NewFailedValidationError(problems)
	}

	return out, nil
}
