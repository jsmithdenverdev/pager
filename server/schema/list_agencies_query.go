package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/jsmithdenverdev/pager/models"
	"github.com/jsmithdenverdev/pager/service"
)

// listAgenciesQuery is the field definition for the agencies query.
var listAgenciesQuery = &graphql.Field{
	Name: "agencies",
	Type: agencyConnectionType,
	Args: graphql.FieldConfigArgument{
		"first": &graphql.ArgumentConfig{
			Type:         graphql.Int,
			DefaultValue: 10,
		},
		"after": &graphql.ArgumentConfig{
			Type: graphql.String,
		},
		"sort": &graphql.ArgumentConfig{
			Type:         agencySortType,
			DefaultValue: service.AgenciesOrderCreatedAsc,
		},
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		type result struct {
			data connection[models.Agency]
			err  error
		}

		ch := make(chan result, 1)

		go func() {
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

			agencies, err := svc.List(pagination)
			ch <- result{data: toConnection(pagination.First, agencies), err: err}
		}()

		// Returning a thunk (a function with a result and error type) tells the
		// graphql engine that this resolver should run concurrently.
		return func() (interface{}, error) {
			r := <-ch
			return r.data, r.err
		}, nil
	},
}
