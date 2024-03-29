package main

import (
	"log/slog"

	"github.com/graphql-go/graphql"
)

func updateAgencyMutation(logger *slog.Logger) *graphql.Field {
	return &graphql.Field{
		Name: "updateAgency",
		Type: graphql.String,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return nil, nil
		},
	}
}
