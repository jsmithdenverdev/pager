package main

import (
	"errors"
	"log/slog"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/graphql-go/graphql"
	"github.com/jmoiron/sqlx"
)

type userStatus string

const (
	userStatusPending  userStatus = "PENDING"
	userStatusActive   userStatus = "ACTIVE"
	userStatusInactive userStatus = "INACTIVE"
)

// user represents a user user in the system.
type user struct {
	auditable
	ID    string `json:"id" db:"id"`
	Email string `json:"email" db:"email"`
	// IdpID is the ID of the user from their identity provider. Typically this
	// comes in the form of a sub claim of an identity token.
	IdpID  string     `json:"idpId" db:"idp_id"`
	Status userStatus `json:"status" db:"status"`
}

var userStatusType = graphql.NewEnum(graphql.EnumConfig{
	Name: "UserStatus",
	Values: graphql.EnumValueConfigMap{
		"PENDING": &graphql.EnumValueConfig{
			Value: userStatusPending,
		},
		"ACTIVE": &graphql.EnumValueConfig{
			Value: userStatusActive,
		},
		"INACTIVE": &graphql.EnumValueConfig{
			Value: userStatusInactive,
		},
	},
})

// userType creates a new graphql object for an agency. The function accepts
// any dependencies needed for field resolvers.
func userType(logger *slog.Logger, agencyType *graphql.Object, authz *authzed.Client, db *sqlx.DB) *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
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
					return p.Source.(user).Created, nil
				},
			},
			"createdBy": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return p.Source.(user).CreatedBy, nil
				},
			},
			"modified": &graphql.Field{
				Type: graphql.DateTime,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return p.Source.(user).Modified, nil
				},
			},
			"modifiedBy": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return p.Source.(user).ModifiedBy, nil
				},
			},
			"agencies": &graphql.Field{
				Type: graphql.NewList(agencyType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					requestContext := p.Context.Value(pagerContextKey{}).(pagerContext)

					// Use the lookupResources data loader to fetch all the agency
					// resources this user has read permission on
					resources, err := requestContext.DataLoaders.lookupResources.Load(p.Context, &v1.LookupResourcesRequest{
						ResourceObjectType: "agency",
						Subject: &v1.SubjectReference{
							Object: &v1.ObjectReference{
								ObjectType: "user",
								ObjectId:   p.Source.(user).IdpID,
							},
						},
						Permission: "read",
					})()

					if err != nil {
						return nil, err
					}

					var agencyKeys []listAgencyKey
					for _, resource := range resources {
						agencyKeys = append(agencyKeys, listAgencyKey{
							key:   resource.ResourceObjectId,
							order: "asc",
						})
					}

					// Use the listAgencies data loader to list all the agencies the user
					// has access to
					results := requestContext.
						DataLoaders.
						listAgencies.
						LoadMany(p.Context, agencyKeys)

					r, errs := results()

					if len(errs) > 0 {
						return nil, errors.Join(errs...)
					}

					return r, nil
				},
			},
		},
	})
}
