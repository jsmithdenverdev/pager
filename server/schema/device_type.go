package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/jsmithdenverdev/pager/models"
	"github.com/jsmithdenverdev/pager/service"
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

// deviceSortType is a graphql enum representing the sort order for devices.
var deviceSortType = graphql.NewEnum(graphql.EnumConfig{
	Name: "DeviceSort",
	Values: graphql.EnumValueConfigMap{
		"CREATED_ASC": &graphql.EnumValueConfig{
			Value: service.DeviceOrderCreatedAsc,
		},
		"CREATED_DESC": &graphql.EnumValueConfig{
			Value: service.DeviceOrderCreatedDesc,
		},
		"MODIFIED_ASC": &graphql.EnumValueConfig{
			Value: service.DeviceOrderModifiedAsc,
		},
		"MODIFIED_DESC": &graphql.EnumValueConfig{
			Value: service.DeviceOrderModifiedDesc,
		},
		"NAME_ASC": &graphql.EnumValueConfig{
			Value: service.DeviceOrderNameAsc,
		},
		"NAME_DESC": &graphql.EnumValueConfig{
			Value: service.DeviceOrderNameDesc,
		},
	},
})

// deviceType is the object definition for a device.
var deviceType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Device",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.ID,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(models.Device).ID, nil
			},
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

// deviceConnectionType represents a relay compliant connection type for
// devices.
var deviceConnectionType = toConnectionType(deviceType)
