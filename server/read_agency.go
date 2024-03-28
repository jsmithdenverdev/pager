package main

import (
	"encoding/json"
	"log/slog"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/graphql-go/graphql"
	"github.com/jmoiron/sqlx"
)

// readAgencyQuery returns the `agency` query.
//
// Allows an authorized user to return the details for an agency given it's ID.
func readAgencyQuery(logger *slog.Logger, types graphTypes, authz *authzed.Client, db *sqlx.DB) *graphql.Field {
	return &graphql.Field{
		Name: "agency",
		Type: toResultType[agency](
			types.agency,
			baseErrorType,
		),
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			id := p.Args["id"].(string)
			requestContext := p.Context.Value(pagerContextKey{}).(pagerContext)

			logger.InfoContext(p.Context, "user", "sub", requestContext.User)

			authzCheck, err := authz.CheckPermission(p.Context, &v1.CheckPermissionRequest{
				Resource: &v1.ObjectReference{
					ObjectType: "agency",
					ObjectId:   id,
				},
				Permission: "read",
				Subject: &v1.SubjectReference{
					Object: &v1.ObjectReference{
						ObjectType: "user",
						ObjectId:   requestContext.User,
					},
				},
			})

			if err != nil {
				return nil, err
			}

			// If you are not authorized to read this agency we return a null value
			if authzCheck.Permissionship != v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION {
				return nil, nil
			}

			var agency agency
			if err := db.QueryRowxContext(
				p.Context,
				"SELECT id, name, created, created_by, modified, modified_by FROM agency WHERE id = $1",
				id,
			).StructScan(&agency); err != nil {
				return nil, err
			}

			bytes, _ := json.MarshalIndent(agency, "", "  ")
			logger.Info(string(bytes))

			return agency, nil
		},
	}
}
