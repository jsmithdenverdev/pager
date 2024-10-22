package handlers

import (
	"context"
	"errors"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
	"github.com/jsmithdenverdev/pager/pkg/authz"
	"github.com/jsmithdenverdev/pager/pkg/codec/apigateway"
	"github.com/jsmithdenverdev/pager/pkg/problemdetail"
	"github.com/jsmithdenverdev/pager/pkg/valid"
	"github.com/jsmithdenverdev/pager/services/agency/internal/config"
)

// createAgencyRequest represents the data required to create a new Agency.
type createAgencyRequest struct {
	Name string `json:"name"`
}

// Valid performs validations on a createAgencyRequest and returns a slice of
// problem structs if issues are encountered.
func (r createAgencyRequest) Valid(ctx context.Context) []valid.Problem {
	var problems []valid.Problem
	if r.Name == "" {
		problems = append(problems, valid.Problem{
			Name:        "name",
			Description: "Name must be at least 1 character",
		})
	}
	return problems
}

type createAgencyResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func CreateAgency(config config.Config, logger *slog.Logger, client *verifiedpermissions.Client) func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		var (
			response      = events.APIGatewayProxyResponse{}
			decoder       = apigateway.NewDecoder[createAgencyRequest]()
			encoder       = apigateway.NewEncoder[createAgencyResponse]()
			pdEncoder     = apigateway.NewEncoder[problemdetail.ProblemDetailer]()
			authClient, _ = authz.RetrieveClientFromContext(ctx)
		)

		// Decode APIGatewayProxyRequest into our request type and validate it
		request, err := decoder.Decode(ctx, event)

		if err != nil {
			// Check if the error was a validation error
			var validErr valid.FailedValidationError
			if errors.As(err, &validErr) {
				// Encode a problem details response for a validation problem, if an error
				// occurs we'll log it. If encode fails a response will still be returned
				// decorated with an internal-server-error problem detail.
				response, encErr := pdEncoder.Encode(ctx, valid.NewProblemDetail(validErr.Problems))
				if encErr != nil {
					logger.ErrorContext(
						ctx,
						"failed to encode response",
						slog.Any("encode error", encErr))
				}

				return response, nil
			}

			// Log the decoding error, this would likely be an error unmarhsaling a
			// request into an expected type.
			logger.ErrorContext(
				ctx,
				"failed to decode request",
				slog.Any("decode error", err))

			// If decoding failed but was not related to validation we'll just encode
			// a generic internal server error and return it.
			response, encErr := pdEncoder.Encode(ctx, problemdetail.New("internal-server-error"))
			if encErr != nil {
				logger.ErrorContext(
					ctx,
					"failed to encode response",
					slog.Any("encode error", encErr))
			}

			return response, nil
		}

		// Check if the user executing the request is authorized to perform the
		// CreateAgency action on the Platform.
		isAuthorized, err := authClient.IsAuthorized(ctx, authz.Resource{
			Type: "pager::Platform",
			ID:   "platform",
		}, authz.Action{
			Type: "pager::Action",
			ID:   "CreateAgency",
		})

		// If an error occurs with authorization log it
		if err != nil {
			logger.ErrorContext(
				ctx,
				"failed authorization check",
				slog.String("error", err.Error()))

			// Encode a generic internal server error and return it. If encoding fails
			// a response will still be returned with the same generic internal server
			// error problem detail attached.
			response, encErr := pdEncoder.Encode(ctx, problemdetail.New("internal-server-error"))
			if encErr != nil {
				logger.ErrorContext(
					ctx,
					"failed to encode response",
					slog.Any("encode error", encErr))
			}

			return response, nil
		}

		if !isAuthorized {
			// If the user is unauthorized encode an unauthorized problem detail
			response, encErr := pdEncoder.Encode(
				ctx,
				authz.NewProblemDetail(authz.NewUnauthorizedError(authz.Resource{
					Type: "pager::Platform",
					ID:   "platform",
				}, authz.Action{
					Type: "pager::Action",
					ID:   "CreateAgency",
				})))

			if encErr != nil {
				logger.ErrorContext(
					ctx,
					"failed to encode response",
					slog.Any("encode error", encErr))
			}

			return response, nil
		}

		response, encErr := encoder.Encode(ctx, createAgencyResponse{
			Name: request.Name,
		})
		if encErr != nil {
			logger.ErrorContext(
				ctx,
				"failed to encode response",
				slog.Any("encode error", encErr))
		}
		return response, nil
	}
}
