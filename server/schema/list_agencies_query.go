package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/jsmithdenverdev/pager/service"
)

// agenciesSortType is a graphql enum representing the sort order for agencies.
var agenciesSortType = graphql.NewEnum(graphql.EnumConfig{
	Name: "AgenciesSort",
	Values: graphql.EnumValueConfigMap{
		"CREATED_ASC": &graphql.EnumValueConfig{
			Value: service.AgenciesOrderCreatedAsc,
		},
		"CREATED_DESC": &graphql.EnumValueConfig{
			Value: service.AgenciesOrderCreatedDesc,
		},
		"MODIFIED_ASC": &graphql.EnumValueConfig{
			Value: service.AgenciesOrderModifiedAsc,
		},
		"MODIFIED_DESC": &graphql.EnumValueConfig{
			Value: service.AgenciesOrderModifiedDesc,
		},
		"NAME_ASC": &graphql.EnumValueConfig{
			Value: service.AgenciesOrderNameAsc,
		},
		"NAME_DESC": &graphql.EnumValueConfig{
			Value: service.AgenciesOrderNameDesc,
		},
	},
})

// listAgenciesQuery is the field definition for the agencies query.
var listAgenciesQuery = &graphql.Field{
	Name: "agencies",
	Type: graphql.NewList(agencyType),
	Args: graphql.FieldConfigArgument{
		"first": &graphql.ArgumentConfig{
			Type:         graphql.Int,
			DefaultValue: 10,
		},
		"after": &graphql.ArgumentConfig{
			Type: graphql.String,
		},
		"sort": &graphql.ArgumentConfig{
			Type:         agenciesSortType,
			DefaultValue: service.AgenciesOrderCreatedAsc,
		},
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		svc := p.Context.Value(service.ContextKeyAgencyService).(*service.AgencyService)
		var (
			argsFirst  = p.Args["first"]
			argsAfter  = p.Args["after"]
			argsOrder  = p.Args["sort"]
			pagination service.AgenciesPagination
		)

		if argsFirst != nil {
			pagination.First = argsFirst.(int)
		}

		if argsAfter != nil {
			pagination.After = argsAfter.(string)
		}

		if argsOrder != nil {
			pagination.Order = argsOrder.(service.AgenciesOrder)
		}

		return svc.List(pagination)
	},
}