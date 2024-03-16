package main

import (
	"github.com/graphql-go/graphql"
)

// agency is the core entity of pager.
//
// An agency represents a real world agency (fire, police, ems, sar, etc.) that
// has a need to recieve pages via push notifications.
//
// Members of an agency are tracked as devices, to which notifications can be
// pushed.
type agency struct {
	// ID is the UUID representing this agency in the pager system.
	ID string `json:"id"`
	// Name is the name of the agency.
	Name string `json:"name"`
}

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
		},
	})
}
