package apigateway

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/jsmithdenverdev/pager/pkg/authz"
	"github.com/jsmithdenverdev/pager/pkg/problemdetail"
	"github.com/jsmithdenverdev/pager/pkg/valid"
)

const (
	problemDetailTypeInternalServerError  = "internal-server-error"
	problemDetailTitleInternalServerError = "Internal Server Error"
)

var (
	rawInternalServerError = fmt.Sprintf(
		`{ "type": "%s", "title": "%s", "status": %d }`,
		problemDetailInternalServerError,
		problemDetailTitleInternalServerError,
		http.StatusInternalServerError,
	)

	problemDetailInternalServerError problemdetail.ProblemDetailer = problemdetail.New(
		problemDetailTypeInternalServerError,
		problemdetail.WithTitle(problemDetailTitleInternalServerError),
	)
)

type Encoder[TIn any] struct {
	logger *slog.Logger
}

type EncoderOption[TIn any] func(*Encoder[TIn])

func WithLogger[TIn any](logger *slog.Logger) EncoderOption[TIn] {
	return func(e *Encoder[TIn]) {
		e.logger = logger
	}
}

func NewEncoder[TIn any](ops ...EncoderOption[TIn]) *Encoder[TIn] {
	encoder := new(Encoder[TIn])
	for _, o := range ops {
		o(encoder)
	}
	return encoder
}

type EncodeOption func(*events.APIGatewayProxyResponse)

func WithStatusCode(code int) EncodeOption {
	return func(apr *events.APIGatewayProxyResponse) {
		apr.StatusCode = code
	}
}

func WithHeader(key, value string) EncodeOption {
	return func(apr *events.APIGatewayProxyResponse) {
		if apr.Headers == nil {
			apr.Headers = make(map[string]string)
		}
		apr.Headers[key] = value
	}
}

func WithHeaders(headers map[string]string) EncodeOption {
	return func(apr *events.APIGatewayProxyResponse) {
		maps.Copy(headers, apr.Headers)
	}
}

func (e *Encoder[TIn]) Encode(ctx context.Context, in TIn, ops ...EncodeOption) (events.APIGatewayProxyResponse, error) {
	response := events.APIGatewayProxyResponse{}

	body, err := json.Marshal(in)
	if err != nil {
		if e.logger != nil {
			e.logger.ErrorContext(ctx, "failed to encode response", "error", err)
		}
		return response, err
	}

	response.Body = string(body)
	for _, o := range ops {
		o(&response)
	}

	return response, nil
}

type ProblemDetailEncoder struct {
	*Encoder[problemdetail.ProblemDetailer]
}

func NewProblemDetailEncoder(opts ...EncoderOption[problemdetail.ProblemDetailer]) *ProblemDetailEncoder {
	encoder := NewEncoder[problemdetail.ProblemDetailer](opts...)
	return &ProblemDetailEncoder{
		Encoder: encoder,
	}
}

func (pde *ProblemDetailEncoder) Encode(ctx context.Context, pd problemdetail.ProblemDetailer, ops ...EncodeOption) events.APIGatewayProxyResponse {
	status := http.StatusInternalServerError
	if pd, ok := pd.(*problemdetail.ProblemDetail); ok {
		status = pd.Status
	}
	ops = append(ops, WithStatusCode(status))
	resp, err := pde.Encoder.Encode(ctx, pd, ops...)
	if err != nil {
		resp = events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       rawInternalServerError,
		}
	}
	return resp
}

func (pde *ProblemDetailEncoder) EncodeInternalServerError(ctx context.Context, ops ...EncodeOption) events.APIGatewayProxyResponse {
	ops = append(ops, WithStatusCode(http.StatusInternalServerError))
	resp, err := pde.Encoder.Encode(ctx, problemdetail.New(problemDetailTypeInternalServerError), ops...)
	if err != nil {
		resp = events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       rawInternalServerError,
		}
	}
	return resp
}

func (pde *ProblemDetailEncoder) EncodeAuthzError(ctx context.Context, authzErr authz.UnauthorizedError, ops ...EncodeOption) events.APIGatewayProxyResponse {
	ops = append(ops, WithStatusCode(http.StatusUnauthorized))
	resp, err := pde.Encoder.Encode(ctx, authz.NewProblemDetail(authzErr), ops...)
	if err != nil {
		resp = events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       rawInternalServerError,
		}
	}
	return resp
}

func (pde *ProblemDetailEncoder) EncodeValidationError(ctx context.Context, validationErr valid.FailedValidationError, ops ...EncodeOption) events.APIGatewayProxyResponse {
	ops = append(ops, WithStatusCode(http.StatusBadRequest))
	resp, err := pde.Encoder.Encode(ctx, valid.NewProblemDetail(validationErr.Problems), ops...)
	if err != nil {
		resp = events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       rawInternalServerError,
		}
	}
	return resp
}
