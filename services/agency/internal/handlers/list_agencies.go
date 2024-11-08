package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
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

func getAgenciesGSI(platformAdmin bool, sort agenciesSort) string {
	if platformAdmin {
		switch sort {
		case agenciesSortCreatedAsc:
			return "type-created-index"
		case agenciesSortCreatedDesc:
			return "type-created-index"
		case agenciesSortModifiedAsc:
			return "type-modified-index"
		case agenciesSortModifiedDesc:
			return "type-modified-index"
		case agenciesSortNameAsc:
			return "type-name-index"
		case agenciesSortNameDesc:
			return "type-name-index"
		default:
			return "type-created-index"
		}
	} else {
		switch sort {
		case agenciesSortCreatedAsc:
			return "idpid-agency_created-index"
		case agenciesSortCreatedDesc:
			return "idpid-agency_created-index"
		case agenciesSortModifiedAsc:
			return "idpid-agency_modified-index"
		case agenciesSortModifiedDesc:
			return "idpid-agency_modified-index"
		case agenciesSortNameAsc:
			return "idpid-name-index"
		case agenciesSortNameDesc:
			return "idpid-name-index"
		default:
			return "idpid-agency_created-index"
		}
	}
}

func ListAgencies(
	config config.Config,
	logger *slog.Logger,
	dynamoClient *dynamodb.Client) func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		var (
			encoder           = apigateway.NewEncoder(apigateway.WithLogger[agenciesResponse](logger))
			errEncoder        = apigateway.NewProblemDetailEncoder(apigateway.WithLogger[problemdetail.ProblemDetailer](logger))
			user, _           = authz.RetrieveUserInfoFromContext(ctx)
			platformAdmin     = false
			first             = 10
			firstStr, firstOk = event.QueryStringParameters["first"]
			after, afterOk    = event.QueryStringParameters["after"]
			sort, sortOk      = event.QueryStringParameters["sort"]
			_, hydrate        = event.QueryStringParameters["hydrate"]
		)

		if firstOk {
			if firstParsed, err := strconv.Atoi(firstStr); err == nil {
				first = firstParsed
			}
		}

		if !sortOk {
			sort = agenciesSortCreatedAsc
		}

		for _, entitlement := range user.Entitlements {
			if entitlement == authz.EntitlementPlatformAdmin {
				platformAdmin = true
			}
		}

		input := &dynamodb.QueryInput{
			TableName: aws.String(config.TableName),
			Limit:     aws.Int32(int32(first)),
		}

		if afterOk && after != "" {
			input.ExclusiveStartKey = map[string]types.AttributeValue{
				"pk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("agency#%s", after),
				},
				"sk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("agency#%s", after),
				},
			}
		}

		input.IndexName = aws.String(getAgenciesGSI(platformAdmin, sort))

		// If sort is ascending
		if sort == agenciesSortCreatedAsc ||
			sort == agenciesSortModifiedAsc ||
			sort == agenciesSortNameAsc {
			input.ScanIndexForward = aws.Bool(true)
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
			// The idpid-agency_created-index can be leveraged to fetch all agencies the
			// current user is a member of, ordered by created date.
			input.IndexName = aws.String("idpid-agency_created-index")
			input.KeyConditionExpression = aws.String("#idpid = :idpid")
			input.ExpressionAttributeNames = map[string]string{"#idpid": "idpid"}
			input.ExpressionAttributeValues = map[string]types.AttributeValue{":idpid": &types.AttributeValueMemberS{Value: user.IPDID}}
		}

		results, err := dynamoClient.Query(ctx, input)
		if err != nil {
			logger.ErrorContext(
				ctx,
				"failed to query dynamodb",
				slog.String("error", err.Error()))

			// If authorization failed encode an internal server error and return it.
			return errEncoder.EncodeInternalServerError(ctx), nil
		}

		response := agenciesResponse{
			Records: make([]agencyResponse, 0),
		}

		// When we list we fetch from a GSI. This results in a partial read. We need
		// to perform a separate hydration step below.
		// TODO: It might be better to project these attributes and prevent the dual
		// read. That just means more attributes to keep in sync. Might not be worth
		// the cost.
		partialAgencies := make([]models.Agency, len(results.Items))
		if err := attributevalue.UnmarshalListOfMaps(results.Items, &partialAgencies); err != nil {
			logger.ErrorContext(
				ctx,
				"failed to unmarshal dynamodb results",
				slog.String("error", err.Error()))

			// If authorization failed encode an internal server error and return it.
			return errEncoder.EncodeInternalServerError(ctx), nil
		}

		// If LastEvaluatedKey is populated that means we have additional results
		// beyond this page.
		if results.LastEvaluatedKey != nil {
			if pkAttr, ok := results.LastEvaluatedKey["pk"].(*types.AttributeValueMemberS); ok {
				response.NextCursor = pkAttr.Value
			}
		}

		if hydrate {
			// Construct a set of keys to hydrate the agency models
			agencyKeys := []map[string]types.AttributeValue{}
			for _, agency := range partialAgencies {
				agencyKeys = append(agencyKeys, map[string]types.AttributeValue{
					"pk": &types.AttributeValueMemberS{
						Value: agency.PK,
					},
					"sk": &types.AttributeValueMemberS{
						Value: agency.SK,
					},
				})
			}

			// Hydrate full agency model
			items, err := dynamoClient.BatchGetItem(ctx, &dynamodb.BatchGetItemInput{
				RequestItems: map[string]types.KeysAndAttributes{
					config.TableName: {
						Keys: agencyKeys,
					},
				},
			})

			if err != nil {
				logger.ErrorContext(
					ctx,
					"failed to batchgetitem dynamodb",
					slog.String("error", err.Error()))

				// If authorization failed encode an internal server error and return it.
				return errEncoder.EncodeInternalServerError(ctx), nil
			}

			hydratedAgencyAttributeValueMaps := items.Responses[config.TableName]

			hydratedAgencies := make([]models.Agency, len(hydratedAgencyAttributeValueMaps))
			if err := attributevalue.UnmarshalListOfMaps(hydratedAgencyAttributeValueMaps, &hydratedAgencies); err != nil {
				logger.ErrorContext(
					ctx,
					"failed to unmarshal dynamodb results",
					slog.String("error", err.Error()))

				// If authorization failed encode an internal server error and return it.
				return errEncoder.EncodeInternalServerError(ctx), nil
			}

			for _, agency := range hydratedAgencies {
				response.Records = append(response.Records, agencyResponse{
					ID:         strings.Split(agency.PK, "#")[1],
					Name:       agency.Name,
					Status:     string(agency.Status),
					Created:    agency.Created,
					CreatedBy:  agency.CreatedBy,
					Modified:   agency.Modified,
					ModifiedBy: agency.ModifiedBy,
					Address:    agency.Address,
					Contact:    agency.Contact,
				})
			}
		} else {
			for _, agency := range partialAgencies {
				response.Records = append(response.Records, agencyResponse{
					ID:       strings.Split(agency.PK, "#")[1],
					Created:  agency.Created,
					Modified: agency.Modified,
				})
			}
		}

		return encoder.Encode(ctx, response, apigateway.WithStatusCode(http.StatusOK))
	}
}
