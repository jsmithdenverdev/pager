package main

import (
	"log/slog"

	"github.com/authzed/authzed-go/v1"
	"github.com/graphql-go/graphql"
	"github.com/jmoiron/sqlx"
)

// readAgencyQuery returns the `userInfo` query.
//
// Allows an authorized user to return their user info including the agencies
// they are a member of. Typically called to retrieve a list of agencies to
// give the user further context.
func userInfoQuery(logger *slog.Logger, types graphTypes, authz *authzed.Client, db *sqlx.DB) *graphql.Field {
	return &graphql.Field{
		Name: "userInfo",
		Type: types.user,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			requestContext := p.Context.Value(pagerContextKey{}).(pagerContext)

			logger.InfoContext(p.Context, "user", "sub", requestContext.User)

			var user user
			if err := db.QueryRowxContext(
				p.Context,
				"SELECT id, email, idp_id, status, created, created_by, modified, modified_by FROM users WHERE idp_id = $1",
				requestContext.User,
			).StructScan(&user); err != nil {
				return nil, err
			}

			return user, nil
		},
	}
}

// readAgencyQuery - TEST returns the `userInfo` query.
//
// Allows an authorized user to return their user info including the agencies
// they are a member of. Typically called to retrieve a list of agencies to
// give the user further context.
func userInfosQuery(logger *slog.Logger, types graphTypes, authz *authzed.Client, db *sqlx.DB) *graphql.Field {
	return &graphql.Field{
		Name: "userInfos",
		Type: graphql.NewList(types.user),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			requestContext := p.Context.Value(pagerContextKey{}).(pagerContext)

			logger.InfoContext(p.Context, "user", "sub", requestContext.User)

			var users []user
			rows, err := db.QueryxContext(
				p.Context,
				"SELECT id, email, idp_id, status, created, created_by, modified, modified_by FROM users")

			if err != nil {
				return users, err
			}

			for rows.Next() {
				var user user
				if err := rows.StructScan(&user); err != nil {
					return users, err
				}
				users = append(users, user)
			}

			return users, nil
		},
	}
}
