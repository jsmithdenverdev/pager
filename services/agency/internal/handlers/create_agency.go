package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
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

func CreateAgency(logger *slog.Logger) func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		var response events.APIGatewayProxyResponse

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
