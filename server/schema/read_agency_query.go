package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/jsmithdenverdev/pager/models"
	"github.com/jsmithdenverdev/pager/service"
)

var readAgencyQuery = &graphql.Field{
	Name: "agency",
	Type: toResultType[models.Agency](
		agencyType,
		baseErrorType,
	),
	Args: graphql.FieldConfigArgument{
		"id": &graphql.ArgumentConfig{
			Type: graphql.String,
		},
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		id := p.Args["id"].(string)
		svc := p.Context.Value(service.ContextKeyAgencyService).(*service.AgencyService)
		return svc.Read(id)
	},
}
