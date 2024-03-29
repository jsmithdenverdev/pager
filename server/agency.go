package main

import (
	"github.com/graphql-go/graphql"
)

type agencyStatus string

const (
	agencyStatusPending  agencyStatus = "PENDING"
	agencyStatusActive   agencyStatus = "ACTIVE"
	agencyStatusInactive agencyStatus = "INACTIVE"
)

// agency is the core entity of pager.
//
// An agency represents a real world agency (fire, police, ems, sar, etc.) that
// has a need to recieve pages via push notifications.
//
// Members of an agency are tracked as devices, to which notifications can be
// pushed.
type agency struct {
	auditable
	// ID is the UUID representing this agency in the pager system.
	ID string `json:"id" db:"id"`
	// Name is the name of the agency.
	Name   string       `json:"name" db:"name"`
	Status agencyStatus `json:"status" db:"status"`
}

var agencyStatusType = graphql.NewEnum(graphql.EnumConfig{
	Name: "AgencyStatus",
	Values: graphql.EnumValueConfigMap{
		"PENDING": &graphql.EnumValueConfig{
			Value: agencyStatusPending,
		},
		"ACTIVE": &graphql.EnumValueConfig{
			Value: agencyStatusActive,
		},
		"INACTIVE": &graphql.EnumValueConfig{
			Value: agencyStatusInactive,
		},
	},
})

// agencyType creates a new graphql object for an agency. The function accepts
// any dependencies needed for field resolvers.
func agencyType() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
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
					return p.Source.(agency).Created, nil
				},
			},
			"createdBy": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return p.Source.(agency).CreatedBy, nil
				},
			},
			"modified": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return p.Source.(agency).Modified, nil
				},
			},
			"modifiedBy": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return p.Source.(agency).ModifiedBy, nil
				},
			},
		},
	})
}
