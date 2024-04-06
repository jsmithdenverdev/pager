package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/jsmithdenverdev/pager/service"
)

// listAgenciesQuery is the field definition for the agencies query.
var listAgenciesQuery = &graphql.Field{
	Name: "agencies",
	Type: graphql.NewList(agencyType),
	Args: graphql.FieldConfigArgument{
		"first": &graphql.ArgumentConfig{
			Type: graphql.Int,
		},
		"after": &graphql.ArgumentConfig{
			Type: graphql.String,
		},
		"sort": &graphql.ArgumentConfig{
			Type: graphql.String,
		},
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		svc := p.Context.Value(service.ContextKeyAgencyService).(*service.AgencyService)
		return svc.List(service.AgenciesPagination{
			First: p.Args["first"].(int),
			After: p.Args["after"].(string),
			Order: service.AgenciesOrderCreatedAsc,
		})
	},
}
