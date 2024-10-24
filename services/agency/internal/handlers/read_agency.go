package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	vptypes "github.com/aws/aws-sdk-go-v2/service/verifiedpermissions/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/jsmithdenverdev/pager/pkg/apigateway"
	"github.com/jsmithdenverdev/pager/pkg/authz"
	"github.com/jsmithdenverdev/pager/pkg/problemdetail"
	"github.com/jsmithdenverdev/pager/services/agency/internal/config"
	"github.com/jsmithdenverdev/pager/services/agency/internal/models"
)

func ReadAgency(
	config config.Config,
	logger *slog.Logger,
	dynamo *dynamodb.Client) func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		var (
			encoder       = apigateway.NewEncoder(apigateway.WithLogger[agencyResponse](logger))
			errEncoder    = apigateway.NewProblemDetailEncoder(apigateway.WithLogger[problemdetail.ProblemDetailer](logger))
			authClient, _ = authz.RetrieveClientFromContext(ctx)
			authzResource = &vptypes.EntityIdentifier{
				EntityType: aws.String("pager::Agency"),
				EntityId:   aws.String(event.PathParameters["agencyid"]),
			}
			authzAction = &vptypes.ActionIdentifier{
				ActionType: aws.String("pager::Action"),
				ActionId:   aws.String("ReadAgency"),
			}
		)

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

		id := event.PathParameters["id"]

		getItemInput := &dynamodb.GetItemInput{
			TableName: aws.String(config.TableName),
			Key: map[string]dbtypes.AttributeValue{
				"pk": &dbtypes.AttributeValueMemberS{
					Value: fmt.Sprintf("agency#%s", id),
				},
				"fk": &dbtypes.AttributeValueMemberS{
					Value: fmt.Sprintf("metadata#%s", id),
				},
			},
		}

		result, err := dynamo.GetItem(ctx, getItemInput)
		if err != nil {
			logger.ErrorContext(
				ctx,
				"failed to get item from dynamodb",
				slog.String("error", err.Error()))

			// If authorization failed encode an internal server error and return it.
			return errEncoder.EncodeInternalServerError(ctx), nil
		}

		if result.Item == nil {
			// TODO: move into apigateway encoder
			pd := problemdetail.New(
				"not-found",
				problemdetail.WithTitle("Not Found"),
				problemdetail.WithInstance(id))
			pd.WriteStatus(http.StatusNotFound)
			return errEncoder.Encode(
				ctx,
				pd,
			), nil
		}

		model := new(models.Agency)

		if err := attributevalue.UnmarshalMap(result.Item, model); err != nil {
			logger.ErrorContext(
				ctx,
				"failed to unmarshal item into agency",
				slog.String("error", err.Error()))

			// If authorization failed encode an internal server error and return it.
			return errEncoder.EncodeInternalServerError(ctx), nil
		}

		response, _ := encoder.Encode(ctx, agencyResponse{
			ID:         id,
			Name:       model.Name,
			Status:     string(model.Status),
			Created:    model.Created,
			CreatedBy:  model.CreatedBy,
			Modified:   model.Modified,
			ModifiedBy: model.ModifiedBy,
		}, apigateway.WithStatusCode(http.StatusOK))

		return response, nil
	}
}
