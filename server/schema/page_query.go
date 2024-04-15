package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/jsmithdenverdev/pager/service"
)

// readPageQuery is the field definition for the page query.
var readPageQuery = &graphql.Field{
	Name: "page",
	Type: pageType,
	Args: graphql.FieldConfigArgument{
		"id": &graphql.ArgumentConfig{
			Type: graphql.String,
		},
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		return func() (interface{}, error) {
			id := p.Args["id"].(string)
			svc := p.Context.Value(service.ContextKeyPageService).(*service.PageService)
			return svc.ReadPage(id)
		}, nil
	},
}
