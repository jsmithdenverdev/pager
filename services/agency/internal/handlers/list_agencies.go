package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/jsmithdenverdev/pager/pkg/apigateway"
	"github.com/jsmithdenverdev/pager/pkg/authz"
	"github.com/jsmithdenverdev/pager/pkg/problemdetail"
	"github.com/jsmithdenverdev/pager/services/agency/internal/config"
	"github.com/jsmithdenverdev/pager/services/agency/internal/models"
)

func ListAgencies(
	config config.Config,
	logger *slog.Logger,
	dynamo *dynamodb.Client) func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		var (
			encoder       = apigateway.NewEncoder(apigateway.WithLogger[agenciesResponse](logger))
			errEncoder    = apigateway.NewProblemDetailEncoder(apigateway.WithLogger[problemdetail.ProblemDetailer](logger))
			user, _       = authz.RetrieveUserInfoFromContext(ctx)
			platformAdmin = false
		)

		for _, entitlement := range user.Entitlements {
			if entitlement == authz.EntitlementPlatformAdmin {
				platformAdmin = true
			}
		}

		input := &dynamodb.QueryInput{
			TableName: aws.String(config.TableName),
		}

		if platformAdmin {
			// The type-created-index can be leveraged to fetch all records of a given
			// type from the database. This allows platform admins to load all
			// agencies.
			input.IndexName = aws.String("type-created-index")
			input.KeyConditionExpression = aws.String("#type = :agencyType")
			input.ExpressionAttributeNames = map[string]string{"#type": "type"}
			input.ExpressionAttributeValues = map[string]types.AttributeValue{":agencyType": &types.AttributeValueMemberS{Value: "AGENCY"}}
		} else {
			// The idpid-agencyId-index can be leveraged to fetch all agencies the
			// current user is a member of.
			input.IndexName = aws.String("idpid-agencyId-index")
			input.KeyConditionExpression = aws.String("#idpid = :idpid")
			input.ExpressionAttributeNames = map[string]string{"#idpid": "idpid"}
			input.ExpressionAttributeValues = map[string]types.AttributeValue{":idpid": &types.AttributeValueMemberS{Value: user.IPDID}}
		}

		results, err := dynamo.Query(ctx, input)
		if err != nil {
			logger.ErrorContext(
				ctx,
				"failed to query dynamodb",
				slog.String("error", err.Error()))

			// If authorization failed encode an internal server error and return it.
			return errEncoder.EncodeInternalServerError(ctx), nil
		}

		response := agenciesResponse{
			Records: make([]struct {
				ID string `json:"id"`
			}, 0),
		}

		for _, item := range results.Items {
			var agency models.Agency
			if err := attributevalue.UnmarshalMap(item, &agency); err != nil {
				logger.ErrorContext(
					ctx,
					"failed to unmarshal dyanmodb row",
					slog.String("error", err.Error()))

				// If authorization failed encode an internal server error and return it.
				return errEncoder.EncodeInternalServerError(ctx), nil
			}

			response.Records = append(response.Records, struct {
				ID string `json:"id"`
			}{
				ID: agency.ID,
			})
		}

		return encoder.Encode(ctx, response, apigateway.WithStatusCode(http.StatusOK))
	}
}
