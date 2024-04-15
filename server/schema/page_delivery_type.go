package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/jsmithdenverdev/pager/models"
	"github.com/jsmithdenverdev/pager/service"
)

// agencyStatusType is a graphql enum representing the status of an agency.
var pageDeliveryStatusType = graphql.NewEnum(graphql.EnumConfig{
	Name: "PageDeliveryStatus",
	Values: graphql.EnumValueConfigMap{
		"PENDING": &graphql.EnumValueConfig{
			Value: models.PageDeliveryStatusPending,
		},
		"DELIVERING": &graphql.EnumValueConfig{
			Value: models.PageDeliveryStatusDelivering,
		},
		"DELIVERED": &graphql.EnumValueConfig{
			Value: models.PageDeliveryStatusDelivered,
		},
		"DELIVERY_FAILED": &graphql.EnumValueConfig{
			Value: models.PageDeliveryStatusDeliveryFailed,
		},
	},
})

// pageSort is a graphql enum representing the sort order for pages.
var pageDeliverySortType = graphql.NewEnum(graphql.EnumConfig{
	Name: "PageDeliverySort",
	Values: graphql.EnumValueConfigMap{
		"CREATED_ASC": &graphql.EnumValueConfig{
			Value: service.PageDeliveryOrderCreatedAsc,
		},
		"CREATED_DESC": &graphql.EnumValueConfig{
			Value: service.PageDeliveryOrderCreatedDesc,
		},
		"MODIFIED_ASC": &graphql.EnumValueConfig{
			Value: service.PageDeliveryOrderModifiedAsc,
		},
		"MODIFIED_DESC": &graphql.EnumValueConfig{
			Value: service.PageDeliveryOrderModifiedDesc,
		},
		"NAME_ASC": &graphql.EnumValueConfig{
			Value: service.PageDeliveryOrderNameAsc,
		},
		"NAME_DESC": &graphql.EnumValueConfig{
			Value: service.PageDeliveryOrderNameDesc,
		},
	},
})

// pageDeliveryType is the object definition for a page delivery.
var pageDeliveryType = graphql.NewObject(graphql.ObjectConfig{
	Name: "PageDelivery",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.ID,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(models.PageDelivery).ID, nil
			},
		},
		"pageId": &graphql.Field{
			Type: graphql.ID,
		},
		"deviceId": &graphql.Field{
			Type: graphql.String,
		},
		"status": &graphql.Field{
			Type: pageDeliveryStatusType,
		},
		"created": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(models.PageDelivery).Created, nil
			},
		},
		"createdBy": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(models.PageDelivery).CreatedBy, nil
			},
		},
		"modified": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(models.PageDelivery).Modified, nil
			},
		},
		"modifiedBy": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(models.PageDelivery).ModifiedBy, nil
			},
		},
	},
})

// pageDeliveryConnectionType  represents a relay compliant connection type for
// page deliveries.
var pageDeliveryConnectionType = toConnectionType(pageDeliveryType)
