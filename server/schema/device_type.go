package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/jsmithdenverdev/pager/models"
)

// deviceStatusType  is a graphql enum representing the status of a device.
var deviceStatusType = graphql.NewEnum(graphql.EnumConfig{
	Name: "DeviceStatus",
	Values: graphql.EnumValueConfigMap{
		"PENDING": &graphql.EnumValueConfig{
			Value: models.DeviceStatusPending,
		},
		"ACTIVE": &graphql.EnumValueConfig{
			Value: models.DeviceStatusActive,
		},
		"INACTIVE": &graphql.EnumValueConfig{
			Value: models.DeviceStatusInactive,
		},
	},
})

// deviceType is the object definition for a device.
var deviceType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Device",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.ID,
		},
		"name": &graphql.Field{
			Type: graphql.String,
		},
		"status": &graphql.Field{
			Type: deviceStatusType,
		},
		"endpoint": &graphql.Field{
			Type: graphql.String,
		},
		"userId": &graphql.Field{
			Type: graphql.String,
		},
		"code": &graphql.Field{
			Type: graphql.String,
		},
		"created": &graphql.Field{
			Type: graphql.DateTime,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(models.Device).Created, nil
			},
		},
		"createdBy": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(models.Device).CreatedBy, nil
			},
		},
		"modified": &graphql.Field{
			Type: graphql.DateTime,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(models.Device).Modified, nil
			},
		},
		"modifiedBy": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(models.Device).ModifiedBy, nil
			},
		},
	},
})
