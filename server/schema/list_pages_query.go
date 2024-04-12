package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/jsmithdenverdev/pager/models"
	"github.com/jsmithdenverdev/pager/service"
)

// pagesSortType is a graphql enum representing the sort order for pages.
var pagesSortType = graphql.NewEnum(graphql.EnumConfig{
	Name: "PagesSort",
	Values: graphql.EnumValueConfigMap{
		"CREATED_ASC": &graphql.EnumValueConfig{
			Value: service.PageOrderCreatedAsc,
		},
		"CREATED_DESC": &graphql.EnumValueConfig{
			Value: service.PageOrderCreatedDesc,
		},
		"MODIFIED_ASC": &graphql.EnumValueConfig{
			Value: service.PageOrderModifiedAsc,
		},
		"MODIFIED_DESC": &graphql.EnumValueConfig{
			Value: service.PageOrderModifiedDesc,
		},
		"NAME_ASC": &graphql.EnumValueConfig{
			Value: service.PageOrderNameAsc,
		},
		"NAME_DESC": &graphql.EnumValueConfig{
			Value: service.PageOrderNameDesc,
		},
	},
})

// listPagesQuery is the field definition for the pages query.
var listPagesQuery = &graphql.Field{
	Name: "pages",
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
			Type:         pagesSortType,
			DefaultValue: service.PageOrderCreatedAsc,
		},
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		type result struct {
			data []models.Page
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
			ch <- result{data: pages, err: err}
		}()

		// Returning a thunk (a function with a result and error type) tells the
		// graphql engine that this resolver should run concurrently.
		return func() (interface{}, error) {
			r := <-ch
			return r.data, r.err
		}, nil
	},
}
