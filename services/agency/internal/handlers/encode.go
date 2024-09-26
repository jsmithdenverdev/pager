package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/jsmithdenverdev/pager/services/agency/internal/authz"
)

// encodeResponse encodes data as a JSON response.
func encodeResponse(
	ctx context.Context,
	response *events.APIGatewayProxyResponse,
	logger *slog.Logger,
	status int,
	data any) {
	response.Headers = make(map[string]string)
	response.Headers["Content-Type"] = "application/json"
	response.StatusCode = status

	b, err := json.Marshal(data)

	if err != nil {
		logger.ErrorContext(
			ctx,
			"[in encodeResponse]: failed to marshal data",
			slog.String("error", err.Error()),
			slog.Any("data", data))

		// If we fail to encode the supplied data, then encode an error response.
		// Note: This does have the potential for causing an infinite loop as
		// encoding the problem details object could result in an error, which
		// would cause us to re-enter this loop indefinitely.
		encodeInternalServerError(ctx, response, logger)
		return
	}

	response.Body = string(b)
}

// encodeValidationError encodes a problemDetails response.
func encodeProblemDetails(
	ctx context.Context,
	response *events.APIGatewayProxyResponse,
	logger *slog.Logger,
	problemDetails problemDetails) {
	encodeResponse(ctx, response, logger, problemDetails.Status, problemDetails)
}

// encodeValidationError encodes a problemDetails response for an HTTP 400 Bad
// Request.
func encodeValidationError(
	ctx context.Context,
	response *events.APIGatewayProxyResponse,
	logger *slog.Logger,
	problems []problem) {
	encodeResponse(ctx, response, logger, http.StatusBadRequest, validationProblemDetails{
		problemDetails: problemDetails{
			Type:   "validation",
			Title:  "Validation Failed",
			Status: http.StatusBadRequest,
			Detail: "The request failed validation",
		},
		Problems: problems,
	})
}

// encodeUnauthorizedError encodes a problemDetails response for an HTTP 401
// Unauthorized.
func encodeUnauthorizedError(
	ctx context.Context,
	response *events.APIGatewayProxyResponse,
	logger *slog.Logger,
	err *authz.AuthzError) {
	encodeProblemDetails(ctx, response, logger, problemDetails{
		Type:     "auth/authorization",
		Title:    "Unauthorized",
		Status:   http.StatusUnauthorized,
		Detail:   string(err.Permission),
		Instance: fmt.Sprintf("%s:%s", err.Resource.Type, err.Resource.ID),
	})
}

// encodeUnauthorizedError encodes a problemDetails response for an  HTTP 500
// Internal Server Error.
func encodeInternalServerError(
	ctx context.Context,
	response *events.APIGatewayProxyResponse,
	logger *slog.Logger) {
	encodeProblemDetails(ctx, response, logger, problemDetails{
		Type:   "unhandled",
		Title:  "Internal Server Error",
		Status: http.StatusInternalServerError,
		Detail: "Please try again",
	})
}
