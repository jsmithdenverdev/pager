package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/jsmithdenverdev/pager/service"
)

// readPageDeliveryQuery is the field definition for the pageDelivery query.
var readPageDeliveryQuery = &graphql.Field{
	Name: "pageDelivery",
	Type: pageDeliveryType,
	Args: graphql.FieldConfigArgument{
		"id": &graphql.ArgumentConfig{
			Type: graphql.String,
		},
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		return func() (interface{}, error) {
			id := p.Args["id"].(string)
			svc := p.Context.Value(service.ContextKeyPageService).(*service.PageService)
			return svc.ReadDelivery(id)
		}, nil
	},
}
