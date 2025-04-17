package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/jsmithdenverdev/pager/pkg/apigateway"
	"github.com/jsmithdenverdev/pager/pkg/authz"
	"github.com/jsmithdenverdev/pager/pkg/problemdetail"
	"github.com/jsmithdenverdev/pager/pkg/valid"
)

// inviteUserRequest represents the data required to invite a new User.
type inviteUserRequest struct {
	Email    string `json:"email"`
	AgencyID string `json:"agencyId"`
	Role     string `json:"role"`
}

type inviteUserResponse struct {
	Email    string `json:"email"`
	AgencyID string `json:"agencyId"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

// Valid performs validations on a createAgencyRequest and returns a slice of
// problem structs if issues are encountered.
func (r inviteUserRequest) Valid(ctx context.Context) []valid.Problem {
	var problems []valid.Problem
	return problems
}

func InviteUser(
	config Config,
	logger *slog.Logger,
	dynamo *dynamodb.Client) func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		var (
			encoder       = apigateway.NewEncoder(apigateway.WithLogger[inviteUserResponse](logger))
			errEncoder    = apigateway.NewProblemDetailEncoder(apigateway.WithLogger[problemdetail.ProblemDetailer](logger))
			authClient, _ = authz.RetrieveClientFromContext(ctx)
			userInfo, _   = authz.RetrieveUserInfoFromContext(ctx)
		)

		// Decode APIGatewayProxyRequest into our request type and validate it
		request, err := apigateway.DecodeValid[inviteUserRequest](ctx, event)

		var (
			authzResource = &types.EntityIdentifier{
				EntityType: aws.String("pager::Agency"),
				EntityId:   aws.String(request.AgencyID),
			}
			authzAction = &types.ActionIdentifier{
				ActionType: aws.String("pager::Action"),
				ActionId:   aws.String("InviteUser"),
			}
		)

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

		model := user{
			Email: request.Email,
			model: model{
				auditable: auditable{
					Created:    time.Now(),
					CreatedBy:  userInfo.IPDID,
					Modified:   time.Now(),
					ModifiedBy: userInfo.IPDID,
				},
			},
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

		response, _ := encoder.Encode(ctx, inviteUserResponse{
			Email: request.Email,
		}, apigateway.WithStatusCode(http.StatusCreated))

		return response, nil
	}
}
