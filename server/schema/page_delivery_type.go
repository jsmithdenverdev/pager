package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/jsmithdenverdev/pager/models"
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

// pageDeliveryType is the object definition for a page delivery.
var pageDeliveryType = graphql.NewObject(graphql.ObjectConfig{
	Name: "PageDelivery",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
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
