package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/jsmithdenverdev/pager/models"
	"github.com/jsmithdenverdev/pager/service"
)

// agencyStatusType is a graphql enum representing the status of an agency.
var agencyStatusType = graphql.NewEnum(graphql.EnumConfig{
	Name: "AgencyStatus",
	Values: graphql.EnumValueConfigMap{
		"PENDING": &graphql.EnumValueConfig{
			Value: models.AgencyStatusPending,
		},
		"ACTIVE": &graphql.EnumValueConfig{
			Value: models.AgencyStatusActive,
		},
		"INACTIVE": &graphql.EnumValueConfig{
			Value: models.AgencyStatusInactive,
		},
	},
})

// agencySortType is a graphql enum representing the sort order for agencies.
var agencySortType = graphql.NewEnum(graphql.EnumConfig{
	Name: "AgencySort",
	Values: graphql.EnumValueConfigMap{
		"CREATED_ASC": &graphql.EnumValueConfig{
			Value: service.AgenciesOrderCreatedAsc,
		},
		"CREATED_DESC": &graphql.EnumValueConfig{
			Value: service.AgenciesOrderCreatedDesc,
		},
		"MODIFIED_ASC": &graphql.EnumValueConfig{
			Value: service.AgenciesOrderModifiedAsc,
		},
		"MODIFIED_DESC": &graphql.EnumValueConfig{
			Value: service.AgenciesOrderModifiedDesc,
		},
		"NAME_ASC": &graphql.EnumValueConfig{
			Value: service.AgenciesOrderNameAsc,
		},
		"NAME_DESC": &graphql.EnumValueConfig{
			Value: service.AgenciesOrderNameDesc,
		},
	},
})

// agencyType is the object definition for an agency.
var agencyType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Agency",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.ID,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(models.Agency).ID, nil
			},
		},
		"name": &graphql.Field{
			Type: graphql.String,
		},
		"status": &graphql.Field{
			Type: agencyStatusType,
		},
		"created": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(models.Agency).Created, nil
			},
		},
		"createdBy": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(models.Agency).CreatedBy, nil
			},
		},
		"modified": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(models.Agency).Modified, nil
			},
		},
		"modifiedBy": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(models.Agency).ModifiedBy, nil
			},
		},
		"devices": &graphql.Field{
			Type: deviceConnectionType,
			Args: graphql.FieldConfigArgument{
				"first": &graphql.ArgumentConfig{
					Type:         graphql.Int,
					DefaultValue: 10,
				},
				"after": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"sort": &graphql.ArgumentConfig{
					Type:         deviceSortType,
					DefaultValue: service.DeviceOrderCreatedAsc,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return func() (interface{}, error) {
					svc := p.Context.Value(service.ContextKeyDeviceService).(*service.DeviceService)
					var (
						argsFirst  = p.Args["first"]
						argsAfter  = p.Args["after"]
						argsOrder  = p.Args["sort"]
						pagination service.DevicePagination
					)

					if argsFirst != nil {
						pagination.First = argsFirst.(int)
					}

					if argsAfter != nil {
						pagination.After = argsAfter.(string)
					}

					if argsOrder != nil {
						pagination.Order = argsOrder.(service.DeviceOrder)
					}
					devices, err := svc.ListDevices(pagination)
					return toConnection(pagination.First, devices), err
				}, nil
			},
		},
		"pages": &graphql.Field{
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
				return func() (interface{}, error) {
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
					return toConnection(pagination.First, pages), err
				}, nil
			},
		},
	},
})

// agencyConnectionType represents a relay compliant connection type for
// agencies.
var agencyConnectionType = toConnectionType(agencyType)
