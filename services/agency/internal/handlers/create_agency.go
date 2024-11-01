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
	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions/types"
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

func CreateAgency(
	config config.Config,
	logger *slog.Logger,
	dynamo *dynamodb.Client) func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		var (
			encoder       = apigateway.NewEncoder(apigateway.WithLogger[agencyResponse](logger))
			errEncoder    = apigateway.NewProblemDetailEncoder(apigateway.WithLogger[problemdetail.ProblemDetailer](logger))
			authClient, _ = authz.RetrieveClientFromContext(ctx)
			userInfo, _   = authz.RetrieveUserInfoFromContext(ctx)
			authzResource = &types.EntityIdentifier{
				EntityType: aws.String("pager::Platform"),
				EntityId:   aws.String("platform"),
			}
			authzAction = &types.ActionIdentifier{
				ActionType: aws.String("pager::Action"),
				ActionId:   aws.String("CreateAgency"),
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
		isAuthorized, err := authClient.IsAuthorized(ctx, authz.IsAuthorizedInput{
			Resource: authzResource,
			Action:   authzAction,
		})

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
				PK:         fmt.Sprintf("agency#%s", id),
				SK:         fmt.Sprintf("metadata#%s", id),
				Created:    time.Now(),
				CreatedBy:  userInfo.IPDID,
				Modified:   time.Now(),
				ModifiedBy: userInfo.IPDID,
			},
			ID:     id,
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

		response, _ := encoder.Encode(ctx, agencyResponse{
			ID:         id,
			Name:       model.Name,
			Status:     string(model.Status),
			Created:    model.Created,
			CreatedBy:  model.CreatedBy,
			Modified:   model.Modified,
			ModifiedBy: model.ModifiedBy,
		}, apigateway.WithStatusCode(http.StatusCreated))

		return response, nil
	}
}
