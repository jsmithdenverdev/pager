package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/jsmithdenverdev/pager/pkg/apigateway"
	"github.com/jsmithdenverdev/pager/pkg/authz"
	"github.com/jsmithdenverdev/pager/pkg/problemdetail"
	"github.com/jsmithdenverdev/pager/pkg/valid"
	"github.com/jsmithdenverdev/pager/services/agency/internal/config"
	"github.com/jsmithdenverdev/pager/services/agency/internal/models"
)

type agenciesSort = string

const (
	agenciesSortCreatedAsc   agenciesSort = "CREATED_ASC"
	agenciesSortCreatedDesc  agenciesSort = "CREATED_DESC"
	agenciesSortModifiedAsc  agenciesSort = "MODIFIED_ASC"
	agenciesSortModifiedDesc agenciesSort = "MODIFIED_DESC"
	agenciesSortNameAsc      agenciesSort = "NAME_ASC"
	agenciesSortNameDesc     agenciesSort = "NAME_DESC"
)

type listAgenciesRequest struct {
	First int          `json:"first"`
	After string       `json:"after"`
	Sort  agenciesSort `json:"sort"`
}

// Valid performs validations on a listAgenciesRequest and returns a slice of
// problem structs if issues are encountered.
// Default values for listAgenciesRequest are also mapped in this method.
func (r listAgenciesRequest) Valid(ctx context.Context) []valid.Problem {
	var problems []valid.Problem
	if r.First == 0 {
		r.First = 10
	}
	if r.Sort == "" {
		r.Sort = agenciesSortCreatedAsc
	}
	return problems
}

func ListAgencies(
	config config.Config,
	logger *slog.Logger,
	dynamoClient *dynamodb.Client) func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		var (
			encoder    = apigateway.NewEncoder(apigateway.WithLogger[agenciesResponse](logger))
			errEncoder = apigateway.NewProblemDetailEncoder(apigateway.WithLogger[problemdetail.ProblemDetailer](logger))
			user, _    = authz.RetrieveUserInfoFromContext(ctx)
		)

		// DynamoDB query for agencies by user
		pk := "pk#user." + user.IPDID
		queryInput := &dynamodb.QueryInput{
			TableName: &config.TableName,
			KeyConditions: map[string]types.Condition{
				"pk": {
					ComparisonOperator: types.ComparisonOperatorEq,
					AttributeValueList: []types.AttributeValue{
						&types.AttributeValueMemberS{Value: pk},
					},
				},
			},
		}
		queryOutput, err := dynamoClient.Query(ctx, queryInput)
		if err != nil {
			logger.Error("failed to query agencies from dynamodb", slog.Any("error", err))
			return errEncoder.EncodeInternalServerError(ctx), nil
		}
		var agencies []models.Agency
		err = attributevalue.UnmarshalListOfMaps(queryOutput.Items, &agencies)
		if err != nil {
			logger.Error("failed to unmarshal agencies from dynamodb", slog.Any("error", err))
			return errEncoder.EncodeInternalServerError(ctx), nil
		}

		var agencyResponses []agencyResponse
		for _, agency := range agencies {
			agencyResponses = append(agencyResponses, agencyResponse{
				ID:         agency.PK,
				Name:       agency.Name,
				Status:     string(agency.Status),
				Created:    agency.Created,
				CreatedBy:  agency.CreatedBy,
				Modified:   agency.Modified,
				ModifiedBy: agency.ModifiedBy,
			})
		}

		return encoder.Encode(ctx, agenciesResponse{Records: agencyResponses}, apigateway.WithStatusCode(http.StatusOK))
	}
}
