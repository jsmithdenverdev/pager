package authz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions/types"
)

type apiGatewayLambdaHandler = func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

type apiGatewayLambdaMiddleware = func(next apiGatewayLambdaHandler) apiGatewayLambdaHandler

func APIGatewayLambdaMiddleware(verifiedPermissionsClient *verifiedpermissions.Client) apiGatewayLambdaMiddleware {
	// encodeResponse encodes data as a JSON response.
	encodeResponse := func(
		ctx context.Context,
		response *events.APIGatewayProxyResponse,
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
	encodeProblemDetails := func(
		ctx context.Context,
		response *events.APIGatewayProxyResponse,
		logger *slog.Logger,
		problemDetails problemDetails) {
		encodeResponse(ctx, response, logger, problemDetails.Status, problemDetails)
	}

	// encodeValidationError encodes a problemDetails response for an HTTP 400 Bad
	// Request.
	encodeValidationError := func(
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
	encodeUnauthorizedError := func(
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
	encodeInternalServerError := func(
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

	return func(next apiGatewayLambdaHandler) apiGatewayLambdaHandler {
		return func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			var userInfo map[string]interface{}
			if err := json.Unmarshal([]byte(event.Headers["x-pager-userinfo"]), &userInfo); err != nil {
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusInternalServerError,
				}, err
			}

			entitlementAttributeValues := []types.AttributeValue{}

			userEntitlements, ok := userInfo["entitlements"].([]interface{})
			if !ok {
				return events.APIGatewayProxyResponse{}, errors.New("failed to convert entitlements to string slice")
			}

			for _, userEntitlement := range userEntitlements {
				entitlementAttributeValues = append(entitlementAttributeValues, &types.AttributeValueMemberString{
					Value: userEntitlement.(string),
				})
			}

			client := New(
				WithVerifiedPermissionsClient(verifiedPermissionsClient),
				WithUserInfo(UserInfo{}))

			ctx = context.WithValue(ctx, "", client)

			return next(ctx, event)
		}
	}
}
