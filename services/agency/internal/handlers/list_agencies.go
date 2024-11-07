package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
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

func getAgenciesGSI(platformAdmin bool, request listAgenciesRequest) string {
	if platformAdmin {
		switch request.Sort {
		default:
			return "type-created-index"
		}
	} else {
		switch request.Sort {
		default:
			return "idpid-created-index"
		}
	}
}

func ListAgencies(
	config config.Config,
	logger *slog.Logger,
	dynamoClient *dynamodb.Client) func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		var (
			encoder       = apigateway.NewEncoder(apigateway.WithLogger[agenciesResponse](logger))
			errEncoder    = apigateway.NewProblemDetailEncoder(apigateway.WithLogger[problemdetail.ProblemDetailer](logger))
			user, _       = authz.RetrieveUserInfoFromContext(ctx)
			platformAdmin = false
		)

		// Decode APIGatewayProxyRequest into our request type and validate it
		request, err := apigateway.DecodeValid[listAgenciesRequest](ctx, event)

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

		for _, entitlement := range user.Entitlements {
			if entitlement == authz.EntitlementPlatformAdmin {
				platformAdmin = true
			}
		}

		input := &dynamodb.QueryInput{
			TableName: aws.String(config.TableName),
			Limit:     aws.Int32(int32(request.First)),
		}

		if request.After != "" {
			input.ExclusiveStartKey = map[string]types.AttributeValue{
				"pk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("agency#%s", request.After),
				},
				"sk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("agency#%s", request.After),
				},
			}
		}

		input.IndexName = aws.String(getAgenciesGSI(platformAdmin, request))

		// If sort is ascending
		if request.Sort == agenciesSortCreatedAsc ||
			request.Sort == agenciesSortModifiedAsc ||
			request.Sort == agenciesSortNameAsc {
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

		agencies := make([]models.Agency, len(results.Items))
		if err := attributevalue.UnmarshalListOfMaps(results.Items, &agencies); err != nil {
			logger.ErrorContext(
				ctx,
				"failed to unmarshal dynamodb results",
				slog.String("error", err.Error()))

			// If authorization failed encode an internal server error and return it.
			return errEncoder.EncodeInternalServerError(ctx), nil
		}

		if results.LastEvaluatedKey != nil {
			if pkAttr, ok := results.LastEvaluatedKey["pk"].(*types.AttributeValueMemberS); ok {
				response.NextCursor = pkAttr.Value
			}
		}

		for _, agency := range agencies {
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

		return encoder.Encode(ctx, response, apigateway.WithStatusCode(http.StatusOK))
	}
}
