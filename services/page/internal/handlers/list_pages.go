package handlers

import (
	"context"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/jsmithdenverdev/pager/pkg/apigateway"
	"github.com/jsmithdenverdev/pager/pkg/problemdetail"
	"github.com/jsmithdenverdev/pager/pkg/valid"
	"github.com/jsmithdenverdev/pager/services/page/internal/config"
)

type pagesSort = string

const (
	pagesSortCreatedAsc   pagesSort = "CREATED_ASC"
	pagesSortCreatedDesc  pagesSort = "CREATED_DESC"
	pagesSortModifiedAsc  pagesSort = "MODIFIED_ASC"
	pagesSortModifiedDesc pagesSort = "MODIFIED_DESC"
	pagesSortNameAsc      pagesSort = "NAME_ASC"
	pagesSortNameDesc     pagesSort = "NAME_DESC"
)

type listAgenciesRequest struct {
	First int       `json:"first"`
	After string    `json:"after"`
	Sort  pagesSort `json:"sort"`
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
		r.Sort = pagesSortCreatedAsc
	}
	return problems
}

func getAgenciesGSI(platformAdmin bool, sort pagesSort) string {
	if platformAdmin {
		switch sort {
		case pagesSortCreatedAsc:
			return "type-created-index"
		case pagesSortCreatedDesc:
			return "type-created-index"
		case pagesSortModifiedAsc:
			return "type-modified-index"
		case pagesSortModifiedDesc:
			return "type-modified-index"
		case pagesSortNameAsc:
			return "type-name-index"
		case pagesSortNameDesc:
			return "type-name-index"
		default:
			return "type-created-index"
		}
	} else {
		switch sort {
		case pagesSortCreatedAsc:
			return "idpid-agency_created-index"
		case pagesSortCreatedDesc:
			return "idpid-agency_created-index"
		case pagesSortModifiedAsc:
			return "idpid-agency_modified-index"
		case pagesSortModifiedDesc:
			return "idpid-agency_modified-index"
		case pagesSortNameAsc:
			return "idpid-name-index"
		case pagesSortNameDesc:
			return "idpid-name-index"
		default:
			return "idpid-agency_created-index"
		}
	}
}

func ListPages(
	config config.Config,
	logger *slog.Logger,
	dynamoClient *dynamodb.Client) func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		var (
			errEncoder = apigateway.NewProblemDetailEncoder(apigateway.WithLogger[problemdetail.ProblemDetailer](logger))
		)

		return errEncoder.EncodeInternalServerError(ctx), nil
	}
}
