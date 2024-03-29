package main

import (
	"log/slog"

	"github.com/graphql-go/graphql"
)

func deleteAgencyMutation(logger *slog.Logger) *graphql.Field {
	return &graphql.Field{
		Name: "deleteAgency",
		Type: graphql.String,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return nil, nil
		},
	}
}
