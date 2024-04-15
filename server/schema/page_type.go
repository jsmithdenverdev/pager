package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/jsmithdenverdev/pager/models"
	"github.com/jsmithdenverdev/pager/service"
)

// pageSort is a graphql enum representing the sort order for pages.
var pageSortType = graphql.NewEnum(graphql.EnumConfig{
	Name: "PageSort",
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

// pageType is the object definition for a page.
var pageType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Page",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.ID,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(models.Page).ID, nil
			},
		},
		"agencyId": &graphql.Field{
			Type: graphql.ID,
		},
		"content": &graphql.Field{
			Type: graphql.String,
		},
		"created": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(models.Page).Created, nil
			},
		},
		"createdBy": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(models.Page).CreatedBy, nil
			},
		},
		"modified": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(models.Page).Modified, nil
			},
		},
		"modifiedBy": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(models.Page).ModifiedBy, nil
			},
		},
		"deliveries": &graphql.Field{
			Type: pageDeliveryConnectionType,
			Args: graphql.FieldConfigArgument{
				"first": &graphql.ArgumentConfig{
					Type:         graphql.Int,
					DefaultValue: 10,
				},
				"after": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"sort": &graphql.ArgumentConfig{
					Type:         pageDeliverySortType,
					DefaultValue: service.PageDeliveryOrderCreatedAsc,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return func() (interface{}, error) {
					svc := p.Context.Value(service.ContextKeyPageService).(*service.PageService)

					var (
						argsFirst  = p.Args["first"]
						argsAfter  = p.Args["after"]
						argsOrder  = p.Args["sort"]
						pagination service.PageDeliveryPagination
					)

					if argsFirst != nil {
						pagination.First = argsFirst.(int)
					}

					if argsAfter != nil {
						pagination.After = argsAfter.(string)
					}

					if argsOrder != nil {
						pagination.Order = argsOrder.(service.PageDeliveryOrder)
					}

					deliveries, err := svc.ListDeliveries(pagination)

					return toConnection(pagination.First, deliveries), err
				}, nil
			},
		},
	},
})

// pageConnectionType represents a relay compliant connection type for pages.
var pageConnectionType = toConnectionType(pageType)
