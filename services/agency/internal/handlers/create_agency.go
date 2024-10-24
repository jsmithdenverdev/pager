package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/google/uuid"
	"github.com/jsmithdenverdev/pager/pkg/apigateway"
	"github.com/jsmithdenverdev/pager/pkg/authz"
	"github.com/jsmithdenverdev/pager/pkg/problemdetail"
	"github.com/jsmithdenverdev/pager/pkg/valid"
	"github.com/jsmithdenverdev/pager/services/agency/internal/config"
	"github.com/jsmithdenverdev/pager/services/agency/internal/models"
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

// createAgencyRequest is the data returned on successful creation of an agency.
type createAgencyResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func CreateAgency(
	config config.Config,
	logger *slog.Logger,
	dynamo *dynamodb.Client) func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		var (
			encoder       = apigateway.NewEncoder(apigateway.WithLogger[createAgencyResponse](logger))
			errEncoder    = apigateway.NewProblemDetailEncoder(apigateway.WithLogger[problemdetail.ProblemDetailer](logger))
			authClient, _ = authz.RetrieveClientFromContext(ctx)
			userInfo, _   = authz.RetrieveUserInfoFromContext(ctx)
			authzResource = authz.Resource{
				Type: "pager::Platform",
				ID:   "platform",
			}
			authzAction = authz.Action{
				Type: "pager::Action",
				ID:   "CreateAgency",
			}
		)

		// Decode APIGatewayProxyRequest into our request type and validate it
		request, err := apigateway.DecodeValid[createAgencyRequest](ctx, event)

		if err != nil {
			// Check if the error was a validation error
			validErr := new(valid.FailedValidationError)
			if errors.As(err, validErr) {
				return errEncoder.EncodeValidationError(ctx, *validErr), nil
			}
			// Log the decoding error, this would likely be an error unmarhsaling a
			// request into an expected type.
			logger.ErrorContext(
				ctx,
				"failed to decode request",
				slog.Any("decode error", err))

			// If decoding failed but was not a validation failure encode an
			// internal server error and return it.
			return errEncoder.EncodeInternalServerError(ctx), nil
		}

		// Check if the user executing the request is authorized to perform the
		// CreateAgency action on the Platform.
		isAuthorized, err := authClient.IsAuthorized(ctx, authzResource, authzAction)

		// If an error occurs with authorization log it
		if err != nil {
			logger.ErrorContext(
				ctx,
				"failed authorization check",
				slog.String("error", err.Error()))

			// If authorization failed encode an internal server error and return it.
			return errEncoder.EncodeInternalServerError(ctx), nil
		}

		if !isAuthorized {
			// Encode and return unauthorized error
			return errEncoder.EncodeAuthzError(ctx, authz.NewUnauthorizedError(authzResource, authzAction)), nil
		}

		id := uuid.NewString()
		model := models.Agency{
			Auditable: models.Auditable{
				ID:         id,
				PK:         fmt.Sprintf("agency#%s", id),
				SK:         fmt.Sprintf("metadata#%s", id),
				Created:    time.Now(),
				CreatedBy:  userInfo.IPDID,
				Modified:   time.Now(),
				ModifiedBy: userInfo.IPDID,
			},
			Name:   request.Name,
			Status: models.AgencyStatusPending,
		}

		dynamoInput, err := attributevalue.MarshalMap(model)
		if err != nil {
			logger.ErrorContext(
				ctx,
				"failed to marshal request to attribute value map",
				slog.String("error", err.Error()))

			// If authorization failed encode an internal server error and return it.
			return errEncoder.EncodeInternalServerError(ctx), nil
		}

		putItemInput := &dynamodb.PutItemInput{
			TableName: aws.String(config.TableName),
			Item:      dynamoInput,
		}

		dynamo.PutItem(ctx, putItemInput)

		response, _ := encoder.Encode(ctx, createAgencyResponse{
			ID:     id,
			Name:   model.Name,
			Status: string(model.Status),
		}, apigateway.WithStatusCode(http.StatusCreated))

		return response, nil
	}
}
