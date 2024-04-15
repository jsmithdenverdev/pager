package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/jsmithdenverdev/pager/models"
	"github.com/jsmithdenverdev/pager/service"
)

// listPagesQuery is the field definition for the pages query.
var listPagesQuery = &graphql.Field{
	Name: "pages",
	Type: pageConnectionType,
	Args: graphql.FieldConfigArgument{
		"first": &graphql.ArgumentConfig{
			Type:         graphql.Int,
			DefaultValue: 10,
		},
		"after": &graphql.ArgumentConfig{
			Type: graphql.String,
		},
		"sort": &graphql.ArgumentConfig{
			Type:         pageSortType,
			DefaultValue: service.PageOrderCreatedAsc,
		},
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		type result struct {
			data connection[models.Page]
			err  error
		}

		ch := make(chan result, 1)

		go func() {
			svc := p.Context.Value(service.ContextKeyPageService).(*service.PageService)
			var (
				argsFirst  = p.Args["first"]
				argsAfter  = p.Args["after"]
				argsOrder  = p.Args["sort"]
				pagination service.PagePagination
			)

			if argsFirst != nil {
				pagination.First = argsFirst.(int)
			}

			if argsAfter != nil {
				pagination.After = argsAfter.(string)
			}

			if argsOrder != nil {
				pagination.Order = argsOrder.(service.PageOrder)
			}

			pages, err := svc.ListPages(pagination)
			ch <- result{data: toConnection(pagination.First, pages), err: err}
		}()

		// Returning a thunk (a function with a result and error type) tells the
		// graphql engine that this resolver should run concurrently.
		return func() (interface{}, error) {
			r := <-ch
			return r.data, r.err
		}, nil
	},
}
