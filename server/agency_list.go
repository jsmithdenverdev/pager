package main

import (
	"context"
	"log/slog"

	"github.com/graph-gophers/dataloader/v7"
	"github.com/jmoiron/sqlx"
)

func newListAgenciesByUserDataloader(logger *slog.Logger, db *sqlx.DB) *dataloader.Loader[string, []agency] {
	return dataloader.NewBatchedLoader(func(ctx context.Context, keys []string) []*dataloader.Result[[]agency] {
		var results []*dataloader.Result[[]agency]

		for _, idpID := range keys {
			logger.InfoContext(ctx, "listAgenciesByUserDataloader", "idpID", idpID)
			var agencies []agency

			rows, err := db.QueryxContext(
				ctx,
				`SELECT a.id, a.name, a.status, a.created, a.created_by, a.modified, a.modified_by
						 FROM agencies a
						 INNER JOIN user_agencies ua ON ua.agency_id = a.id
						 INNER JOIN users u ON u.id = ua.user_id
						 WHERE u.idp_id = $1`,
				idpID)

			if err != nil {
				results = append(results, &dataloader.Result[[]agency]{
					Error: err,
				})
			}

			for rows.Next() {
				var a agency
				if err := rows.StructScan(&a); err != nil {
					results = append(results, &dataloader.Result[[]agency]{
						Error: err,
					})
				}
				agencies = append(agencies, a)
			}
			results = append(results, &dataloader.Result[[]agency]{
				Data: agencies,
			})
		}

		return results
	})
}
