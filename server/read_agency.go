package main

import (
	"log/slog"

	"github.com/graphql-go/graphql"
)

func readAgencyQuery(logger *slog.Logger) *graphql.Field {
	return &graphql.Field{
		Name: "agency",
		Type: graphql.String,
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return nil, nil
		},
	}
}
