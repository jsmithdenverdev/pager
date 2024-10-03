package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions/types"
	"github.com/jsmithdenverdev/pager/services/agency/internal/authz"
	"github.com/jsmithdenverdev/pager/services/agency/internal/config"
	"github.com/jsmithdenverdev/pager/services/agency/internal/models"
)

// createAgencyRequest represents the data required to create a new Agency.
type createAgencyRequest struct {
	Name string `json:"name"`
}

// Valid performs validations on a createAgencyRequest and returns a slice of
// problem structs if issues are encountered.
func (r createAgencyRequest) Valid(ctx context.Context) []problem {
	var problems []problem
	if r.Name == "" {
		problems = append(problems, problem{
			Name:        "name",
			Description: "Name must be at least 1 character",
		})
	}
	return problems
}

// MapTo maps a createAgencyRequest to a models.Agency.
func (r createAgencyRequest) MapTo() models.Agency {
	var m models.Agency
	m.Name = r.Name
	return m
}

func CreateAgency(config config.Config, logger *slog.Logger, client *verifiedpermissions.Client) func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		var response events.APIGatewayProxyResponse

		authzResult, err := client.IsAuthorized(ctx, &verifiedpermissions.IsAuthorizedInput{
			PolicyStoreId: aws.String(config.PolicyStoreID),
			// TODO: Populate with action
			Action: &types.ActionIdentifier{},
		})

		if err != nil {
			logger.ErrorContext(
				ctx,
				"[in handlers.CreateAgency] failed to check authorization",
				slog.String("error", err.Error()))

			encodeInternalServerError(ctx, &response, logger)
			return response, nil
		}

		if authzResult.Decision != types.DecisionAllow {
			authzErr := authz.NewAuthzError(authz.PermissionCreateAgency, authz.Resource{
				Type: "Platform",
				ID:   "platform",
			})
			encodeUnauthorizedError(ctx, &response, logger, &authzErr)
			return response, nil
		}

		req, problems, err := decodeValid[createAgencyRequest](ctx, event)

		if err != nil {
			if len(problems) > 0 {
				encodeValidationError(ctx, &response, logger, problems)
				return response, nil
			} else {
				logger.ErrorContext(
					ctx,
					"[in handlers.CreateAgency] failed to decode request",
					slog.String("error", err.Error()))

				encodeInternalServerError(ctx, &response, logger)
				return response, nil
			}
		}

		b, _ := json.Marshal(req)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       string(b),
		}, nil
	}
}
