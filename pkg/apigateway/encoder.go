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

// Encoder is a generic type responsible for encoding an input of type TIn
// into an API Gateway Proxy Response. It optionally includes a logger for
// logging errors during encoding.
type Encoder[TIn any] struct {
	logger *slog.Logger
}

// EncoderOption represents a functional option for configuring an Encoder.
type EncoderOption[TIn any] func(*Encoder[TIn])

// WithLogger returns an EncoderOption that sets a logger for the Encoder.
func WithLogger[TIn any](logger *slog.Logger) EncoderOption[TIn] {
	return func(e *Encoder[TIn]) {
		e.logger = logger
	}
}

// NewEncoder creates a new Encoder, applying any provided EncoderOptions.
func NewEncoder[TIn any](ops ...EncoderOption[TIn]) *Encoder[TIn] {
	encoder := new(Encoder[TIn])
	for _, o := range ops {
		o(encoder)
	}
	return encoder
}

// EncodeOption represents a functional option for modifying an
// APIGatewayProxyResponse during encoding.
type EncodeOption func(*events.APIGatewayProxyResponse)

// WithStatusCode returns an EncodeOption that sets the status code
// for the API Gateway Proxy Response.
func WithStatusCode(code int) EncodeOption {
	return func(apr *events.APIGatewayProxyResponse) {
		apr.StatusCode = code
	}
}

// WithHeader returns an EncodeOption that adds a single header to the
// API Gateway Proxy Response.
func WithHeader(key, value string) EncodeOption {
	return func(apr *events.APIGatewayProxyResponse) {
		if apr.Headers == nil {
			apr.Headers = make(map[string]string)
		}
		apr.Headers[key] = value
	}
}

// WithHeaders returns an EncodeOption that adds multiple headers to the
// API Gateway Proxy Response.
func WithHeaders(headers map[string]string) EncodeOption {
	return func(apr *events.APIGatewayProxyResponse) {
		maps.Copy(headers, apr.Headers)
	}
}

// Encode encodes the input of type TIn into an API Gateway Proxy Response
// using the provided EncodeOptions, and logs any errors if a logger is set.
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

// ProblemDetailEncoder is a specialized encoder for encoding problem details,
// leveraging the generic Encoder with problemdetail.ProblemDetailer as the input type.
type ProblemDetailEncoder struct {
	*Encoder[problemdetail.ProblemDetailer]
}

// NewProblemDetailEncoder creates a new ProblemDetailEncoder with the provided options.
func NewProblemDetailEncoder(opts ...EncoderOption[problemdetail.ProblemDetailer]) *ProblemDetailEncoder {
	encoder := NewEncoder[problemdetail.ProblemDetailer](opts...)
	return &ProblemDetailEncoder{
		Encoder: encoder,
	}
}

// Encode encodes a problemdetail.ProblemDetailer into an API Gateway Proxy Response.
// It determines the status code based on the problem detail object, falling back to 500.
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

// EncodeInternalServerError encodes a generic internal server error as a problem detail
// into an API Gateway Proxy Response, setting the status code to 500.
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

// EncodeAuthzError encodes an authorization error as a problem detail
// into an API Gateway Proxy Response, setting the status code to 401.
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

// EncodeValidationError encodes a validation error as a problem detail
// into an API Gateway Proxy Response, setting the status code to 400.
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
