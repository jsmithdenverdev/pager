package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/graph-gophers/dataloader/v7"
	"github.com/graphql-go/graphql"
	"github.com/jmoiron/sqlx"
)

type listAgencyKey struct {
	order string
	key   string
}

func newListAgenciesDataLoader(logger *slog.Logger, db *sqlx.DB) *dataloader.Loader[listAgencyKey, agency] {
	return dataloader.NewBatchedLoader(func(ctx context.Context, keys []listAgencyKey) []*dataloader.Result[agency] {
		var results []*dataloader.Result[agency]

		var ids []string
		for _, key := range keys {
			ids = append(ids, key.key)
		}

		q := fmt.Sprintf(`SELECT id, name, status, created, created_by, modified, modified_by
						 FROM agencies
						 WHERE id IN (?)
						 ORDER BY created %s`, keys[0].order)

		query, args, err := sqlx.In(
			q,
			ids)

		if err != nil {
			results = append(results, &dataloader.Result[agency]{
				Error: err,
			})
		}

		query = db.Rebind(query)

		rows, err := db.QueryxContext(
			ctx,
			query,
			args...,
		)

		if err != nil {
			results = append(results, &dataloader.Result[agency]{
				Error: err,
			})
		}

		for rows.Next() {
			var a agency
			if err := rows.StructScan(&a); err != nil {
				results = append(results, &dataloader.Result[agency]{
					Error: err,
				})
			}
			results = append(results, &dataloader.Result[agency]{
				Data: a,
			})
		}

		return results
	}, dataloader.WithCache[listAgencyKey, agency](&dataloader.NoCache[listAgencyKey, agency]{}))
}

// listAgenciesQuery returns the `agencies` query.
//
// Allows an authorized user to return a list of agencies they have access to.
func listAgenciesQuery(types graphTypes) *graphql.Field {
	return &graphql.Field{
		Name: "agencies",
		Type: graphql.NewList(types.agency),
		Args: graphql.FieldConfigArgument{
			"sort": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			requestContext := p.Context.Value(pagerContextKey{}).(pagerContext)
			resources, err := requestContext.DataLoaders.lookupResources.Load(p.Context, &v1.LookupResourcesRequest{
				ResourceObjectType: "agency",
				Subject: &v1.SubjectReference{
					Object: &v1.ObjectReference{
						ObjectType: "user",
						ObjectId:   requestContext.User,
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
					order: p.Args["sort"].(string),
				})
			}

			results, errs := requestContext.
				DataLoaders.
				listAgencies.
				LoadMany(p.Context, agencyKeys)()

			return results, errors.Join(errs...)
		},
	}
}
