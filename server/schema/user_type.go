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

// userSortType is a graphql enum representing the sort order for users.
var userSortType = graphql.NewEnum(graphql.EnumConfig{
	Name: "UserSort",
	Values: graphql.EnumValueConfigMap{
		"CREATED_ASC": &graphql.EnumValueConfig{
			Value: service.UsersOrderCreatedAsc,
		},
		"CREATED_DESC": &graphql.EnumValueConfig{
			Value: service.UsersOrderCreatedDesc,
		},
		"MODIFIED_ASC": &graphql.EnumValueConfig{
			Value: service.UsersOrderModifiedAsc,
		},
		"MODIFIED_DESC": &graphql.EnumValueConfig{
			Value: service.UsersOrderModifiedDesc,
		},
		"NAME_ASC": &graphql.EnumValueConfig{
			Value: service.UsersOrderNameAsc,
		},
		"NAME_DESC": &graphql.EnumValueConfig{
			Value: service.UsersOrderNameDesc,
		},
	},
})

// userType is the object definition for a user.
var userType = graphql.NewObject(graphql.ObjectConfig{
	Name: "User",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.ID,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(models.User).ID, nil
			},
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
			// We're defining a new type instead of using the agencyConnectionType,
			// this prevents an cyclic initialization error that would arrise when
			// agencyType depends on userType which depends on agencyType.
			Type: userAgencyConnectionType,
			Args: graphql.FieldConfigArgument{
				"first": &graphql.ArgumentConfig{
					Type:         graphql.Int,
					DefaultValue: 10,
				},
				"after": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"sort": &graphql.ArgumentConfig{
					Type:         agencySortType,
					DefaultValue: service.AgenciesOrderCreatedAsc,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				type result struct {
					data connection[models.Agency]
					err  error
				}

				ch := make(chan result, 1)

				go func() {
					svc := p.Context.Value(service.ContextKeyAgencyService).(*service.AgencyService)
					var (
						argsFirst  = p.Args["first"]
						argsAfter  = p.Args["after"]
						argsOrder  = p.Args["sort"]
						pagination service.AgenciesPagination
					)

					if argsFirst != nil {
						pagination.First = argsFirst.(int)
					}

					if argsAfter != nil {
						pagination.After = argsAfter.(string)
					}

					if argsOrder != nil {
						pagination.Order = argsOrder.(service.AgenciesOrder)
					}

					agencies, err := svc.ListAgencies(pagination)
					ch <- result{data: toConnection(pagination.First, agencies), err: err}
				}()

				// Returning a thunk (a function with a result and error type) tells the
				// graphql engine that this resolver should run concurrently.
				return func() (interface{}, error) {
					r := <-ch
					return r.data, r.err
				}, nil
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

					pagination.Filter.UserID = p.Source.(models.User).ID

					devices, err := svc.ListDevices(pagination)
					return toConnection(pagination.First, devices), err
				}, nil
			},
		},
		"roles": &graphql.Field{
			Type: graphql.NewList(userRoleType),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return func() (interface{}, error) {
					svc := p.Context.Value(service.ContextKeyUserService).(*service.UserService)
					return svc.Roles()
				}, nil
			},
		},
	},
})

// userConnectionType represents a relay compliant connection type for
// users.
var userConnectionType = toConnectionType(userType)
