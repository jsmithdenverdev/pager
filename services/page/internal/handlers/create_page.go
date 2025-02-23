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
	"github.com/jsmithdenverdev/pager/services/page/internal/config"
	"github.com/jsmithdenverdev/pager/services/page/internal/models"
)

// createPageRequest represents the data required to create a new Page.
type createPageRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Location    struct {
		Type models.LocationType `json:"type"`
		Data string              `json:"data"`
	} `json:"location"`
}

// Valid performs validations on a createPageRequest and returns a slice of
// problem structs if issues are encountered.
func (r createPageRequest) Valid(ctx context.Context) []valid.Problem {
	var problems []valid.Problem
	return problems
}

func CreateAgency(
	config config.Config,
	logger *slog.Logger,
	dynamo *dynamodb.Client) func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		var (
			encoder       = apigateway.NewEncoder(apigateway.WithLogger[pageResponse](logger))
			errEncoder    = apigateway.NewProblemDetailEncoder(apigateway.WithLogger[problemdetail.ProblemDetailer](logger))
			authClient, _ = authz.RetrieveClientFromContext(ctx)
			userInfo, _   = authz.RetrieveUserInfoFromContext(ctx)
			agencyId, _   = event.Headers["agencyid"]
			authzResource = &types.EntityIdentifier{
				EntityType: aws.String("pager::Agency"),
				EntityId:   aws.String(agencyId),
			}
			authzAction = &types.ActionIdentifier{
				ActionType: aws.String("pager::Action"),
				ActionId:   aws.String("CreatePage"),
			}
		)

		// Decode APIGatewayProxyRequest into our request type and validate it
		request, err := apigateway.DecodeValid[createPageRequest](ctx, event)

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
		// CreatePage action on the Agency.
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

		id := uuid.New().String()

		page := models.Page{
			Auditable: models.Auditable{
				PK:         fmt.Sprintf("page#%s", id),
				SK:         fmt.Sprintf("page#%s", id),
				Created:    time.Now(),
				CreatedBy:  userInfo.IPDID,
				Modified:   time.Now(),
				ModifiedBy: userInfo.IPDID,
			},
			Title:       request.Title,
			Description: request.Description,
			Location: models.Location{
				Type: request.Location.Type,
				Data: request.Location.Data,
			},
		}

		// TODO: Create Page, PageAgency
		dynamoInput, err := attributevalue.MarshalMap(page)
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

		_, err = dynamo.PutItem(ctx, putItemInput)
		if err != nil {
			// Log the dynamo error
			logger.ErrorContext(
				ctx,
				"failed to put item in dynamo",
				slog.Any("put item error", err))

			response := errEncoder.EncodeInternalServerError(ctx)
			return response, nil
		}

		// TODO: Perhaps this should be in a mapper function? This feels brittle
		pageResponse := new(pageResponse)
		pageResponse.ID = id
		pageResponse.Title = page.Title
		pageResponse.Description = page.Description
		pageResponse.Created = page.Created
		pageResponse.CreatedBy = page.CreatedBy
		pageResponse.Modified = page.Modified
		pageResponse.ModifiedBy = page.CreatedBy
		pageResponse.Location.Type = page.Location.Type
		pageResponse.Location.Data = page.Location.Data

		response, _ := encoder.Encode(
			ctx,
			*pageResponse,
			apigateway.WithStatusCode(http.StatusCreated))

		return response, nil
	}
}
