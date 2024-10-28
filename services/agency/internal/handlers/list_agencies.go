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
			encoder    = apigateway.NewEncoder(apigateway.WithLogger[agenciesResponse](logger))
			errEncoder = apigateway.NewProblemDetailEncoder(apigateway.WithLogger[problemdetail.ProblemDetailer](logger))
			user, _    = authz.RetrieveUserInfoFromContext(ctx)
		)

		agencyKeys := make([]map[string]dbtypes.AttributeValue, 0)

		for agencyId := range user.Agencies {
			agencyKeys = append(agencyKeys, map[string]dbtypes.AttributeValue{
				"pk": &dbtypes.AttributeValueMemberS{
					Value: fmt.Sprintf("agency#%s", agencyId),
				},
				"sk": &dbtypes.AttributeValueMemberS{
					Value: fmt.Sprintf("metadata#%s", agencyId),
				},
			})
		}

		batchGetItemInput := &dynamodb.BatchGetItemInput{
			RequestItems: map[string]dbtypes.KeysAndAttributes{
				config.TableName: {
					Keys: agencyKeys,
				},
			},
		}

		results, err := dynamo.BatchGetItem(ctx, batchGetItemInput)
		if err != nil {
			logger.ErrorContext(
				ctx,
				"failed to get item from dynamodb",
				slog.String("error", err.Error()))

			// If authorization failed encode an internal server error and return it.
			return errEncoder.EncodeInternalServerError(ctx), nil
		}

		if records, ok := results.Responses[config.TableName]; ok {
			agencies := make([]agencyResponse, 0)
			agenciesResponse := agenciesResponse{}
			for _, record := range records {
				model := new(models.Agency)

				if err := attributevalue.UnmarshalMap(record, model); err != nil {
					logger.ErrorContext(
						ctx,
						"failed to unmarshal item into agency",
						slog.String("error", err.Error()))

					// If authorization failed encode an internal server error and return it.
					return errEncoder.EncodeInternalServerError(ctx), nil
				}

				agencies = append(agencies, agencyResponse{
					ID:         model.ID,
					Name:       model.Name,
					Status:     string(model.Status),
					Created:    model.Created,
					CreatedBy:  model.CreatedBy,
					Modified:   model.Modified,
					ModifiedBy: model.ModifiedBy,
				})

			}
			agenciesResponse.Records = agencies

			response, _ := encoder.Encode(ctx, agenciesResponse, apigateway.WithStatusCode(http.StatusOK))
			return response, nil
		} else {
			response, _ := encoder.Encode(
				ctx,
				agenciesResponse{
					Records: make([]agencyResponse, 0),
				},
				apigateway.WithStatusCode(http.StatusOK))

			return response, nil
		}

	}
}
