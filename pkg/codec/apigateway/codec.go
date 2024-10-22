package apigateway

import (
	"context"
	"encoding/json"
	"maps"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/jsmithdenverdev/pager/pkg/codec"
	"github.com/jsmithdenverdev/pager/pkg/valid"
)

func NewDecoder[TOut valid.Validator]() codec.Decoder[events.APIGatewayProxyRequest, TOut] {
	return codec.DecoderFunc[events.APIGatewayProxyRequest, TOut](
		func(ctx context.Context, in events.APIGatewayProxyRequest) (TOut, error) {
			var out TOut
			if err := json.Unmarshal([]byte(in.Body), &out); err != nil {
				return out, err
			}

			if problems := out.Valid(ctx); len(problems) > 0 {
				return *new(TOut), valid.NewFailedValidationError(problems)
			}

			return *new(TOut), nil
		})
}

func NewEncoder[TIn any]() codec.Encoder[TIn, events.APIGatewayProxyResponse] {
	return codec.EncoderFunc[TIn, events.APIGatewayProxyResponse](func(ctx context.Context, in TIn, opts ...codec.EncoderOption[events.APIGatewayProxyResponse]) (events.APIGatewayProxyResponse, error) {
		response := events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
		}

		b, err := json.Marshal(in)
		if err != nil {
			return events.APIGatewayProxyResponse{}, err
		}

		response.Body = string(b)

		for _, op := range opts {
			op(&response)
		}

		return response, nil
	})
}

func WithStatusCode(code int) codec.EncoderOption[events.APIGatewayProxyResponse] {
	return func(out *events.APIGatewayProxyResponse) {
		out.StatusCode = code
	}
}

func WithHeader(key, value string) codec.EncoderOption[events.APIGatewayProxyResponse] {
	return func(out *events.APIGatewayProxyResponse) {
		if out.Headers == nil {
			out.Headers = make(map[string]string)
		}
		out.Headers[key] = value
	}
}

func WithHeaders(headers map[string]string) codec.EncoderOption[events.APIGatewayProxyResponse] {
	return func(out *events.APIGatewayProxyResponse) {
		if out.Headers == nil {
			out.Headers = make(map[string]string)
		}
		maps.Copy(out.Headers, headers)
	}
}
