package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/jsmithdenverdev/pager/models"
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

// agencyType creates a new graphql object for an agency. The function accepts
// any dependencies needed for field resolvers.
var agencyType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Agency",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
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
	},
})
