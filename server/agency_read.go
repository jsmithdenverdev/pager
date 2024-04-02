package main

import (
	"context"
	"log/slog"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/graph-gophers/dataloader/v7"
	"github.com/graphql-go/graphql"
	"github.com/jmoiron/sqlx"
)

func newReadAgencyDataLoader(db *sqlx.DB) *dataloader.Loader[string, agency] {
	return dataloader.NewBatchedLoader(func(ctx context.Context, keys []string) []*dataloader.Result[agency] {
		var results []*dataloader.Result[agency]
		for _, id := range keys {
			var a agency
			err := db.QueryRowxContext(
				ctx,
				`SELECT id, name, status, created, created_by, modified, modified_by
				 FROM agencies 
				 WHERE id = $1`,
				id,
			).StructScan(&a)

			results = append(results, &dataloader.Result[agency]{
				Data:  a,
				Error: err,
			})
		}
		return results
	})
}

// readAgencyQuery returns the `agency` query.
//
// Allows an authorized user to return the details for an agency given it's ID.
func readAgencyQuery(logger *slog.Logger, types graphTypes) *graphql.Field {
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

			authzCheck, err := requestContext.
				DataLoaders.
				checkPermission.
				Load(p.Context, &v1.CheckPermissionRequest{
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
				})()

			if err != nil {
				return nil, err
			}

			// If you are not authorized to read this agency we return a null value
			if authzCheck.Permissionship != v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION {
				return nil, nil
			}

			return requestContext.
				DataLoaders.
				readAgency.
				Load(p.Context, id)()
		},
	}
}
