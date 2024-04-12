package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/jsmithdenverdev/pager/models"
	"github.com/jsmithdenverdev/pager/service"
)

// userStatusType is a graphql enum representing the status of a user.
var userStatusType = graphql.NewEnum(graphql.EnumConfig{
	Name: "UserStatus",
	Values: graphql.EnumValueConfigMap{
		"PENDING": &graphql.EnumValueConfig{
			Value: models.UserStatusPending,
		},
		"ACTIVE": &graphql.EnumValueConfig{
			Value: models.UserStatusActive,
		},
		"INACTIVE": &graphql.EnumValueConfig{
			Value: models.UserStatusInactive,
		},
	},
})

// userType is the object definition for a user.
var userType = graphql.NewObject(graphql.ObjectConfig{
	Name: "User",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"email": &graphql.Field{
			Type: graphql.String,
		},
		"idpId": &graphql.Field{
			Type: graphql.String,
		},
		"status": &graphql.Field{
			Type: userStatusType,
		},
		"created": &graphql.Field{
			Type: graphql.DateTime,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(models.User).Created, nil
			},
		},
		"createdBy": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(models.User).CreatedBy, nil
			},
		},
		"modified": &graphql.Field{
			Type: graphql.DateTime,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(models.User).Modified, nil
			},
		},
		"modifiedBy": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(models.User).ModifiedBy, nil
			},
		},
		"agencies": &graphql.Field{
			Type: graphql.NewList(agencyType),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return func() (interface{}, error) {
					svc := p.Context.Value(service.ContextKeyAgencyService).(*service.AgencyService)
					return svc.List(service.AgenciesPagination{
						First: 10,
						Order: service.AgenciesOrderCreatedAsc,
					})
				}, nil
			},
		},
		"devices": &graphql.Field{
			Type: graphql.NewList(deviceType),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return func() (interface{}, error) {
					svc := p.Context.Value(service.ContextKeyDeviceService).(*service.DeviceService)
					return svc.ListDevices(service.DevicePagination{
						First: 10,
						Order: service.DeviceOrderCreatedAsc,
						Filter: struct {
							AgencyID string
							UserID   string
						}{
							UserID: p.Source.(models.User).ID,
						},
					})
				}, nil
			},
		},
	},
})
