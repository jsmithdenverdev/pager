package service

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/graph-gophers/dataloader/v7"
	"github.com/jmoiron/sqlx"
	"github.com/jsmithdenverdev/pager/authz"
	"github.com/jsmithdenverdev/pager/models"
)

type AgenciesOrder string

const (
	AgenciesOrderNameAsc AgenciesOrder = "NAME_ASC"
)

type AgenciesPagination struct {
	First int
	Order AgenciesOrder
}

func newListAgenciesDataLoader(authclient *authz.Client, db *sqlx.DB) *dataloader.Loader[AgenciesPagination, []models.Agency] {
	return dataloader.NewBatchedLoader(
		func(ctx context.Context, paginations []AgenciesPagination) []*dataloader.Result[[]models.Agency] {
			results := make([]*dataloader.Result[[]models.Agency], len(paginations))
			// Give me the list of agencies this user has the read permission on
			ids, err := authclient.List("read", authz.Resource{Type: "agency"})
			// We're not actually using pagination settings yet, I just want to get
			// this working through a loader to start.
			for i, _ := range paginations {
				if err != nil {
					results[i] = &dataloader.Result[[]models.Agency]{
						Error: err,
					}
					continue
				}

				// Query the agencies
				query, args, err := sqlx.In(
					`SELECT id, name, status, created, created_by, modified, modified_by
					 FROM agencies
					 WHERE id IN (?)
					 ORDER BY created DESC`,
					ids)

				if err != nil {
					results[i] = &dataloader.Result[[]models.Agency]{
						Error: err,
					}
					continue
				}

				query = db.Rebind(query)

				rows, err := db.QueryxContext(
					ctx,
					query,
					args...,
				)

				if err != nil {
					results[i] = &dataloader.Result[[]models.Agency]{
						Error: err,
					}
					continue
				}

				var agencies []models.Agency
				for rows.Next() {
					var a models.Agency
					if err := rows.StructScan(&a); err != nil {
						results[i] = &dataloader.Result[[]models.Agency]{
							Error: err,
						}
						if err := rows.Close(); err != nil {
							results[i] = &dataloader.Result[[]models.Agency]{
								Error: err,
							}
						}
						// Here we break instead of continue. This closes the db connection
						// and considers this entire query set a failure.
						break
					}
					agencies = append(agencies, a)
				}

				results[i] = &dataloader.Result[[]models.Agency]{
					Data: agencies,
				}
			}
			return results
		})
}

func newReadAgencyDataLoader(authclient *authz.Client, db *sqlx.DB) *dataloader.Loader[string, models.Agency] {
	return dataloader.NewBatchedLoader(
		func(ctx context.Context, keys []string) []*dataloader.Result[models.Agency] {
			results := make([]*dataloader.Result[models.Agency], len(keys))
			var resources []authz.Resource

			// Build up a collection of resources for a batch authorization check. We
			// do this because this dataloader may be called multiple times to fully
			// resolve a particular query. This allows us to coalesce the full set of
			// IDs from all calls to `.Load` into a single authorization query.
			for _, key := range keys {
				resources = append(resources, authz.Resource{Type: "agency", ID: key})
			}

			// BatchAuthorize returns a result set the same length as the input set.
			// Each result in the set is a boolean that can be used to determine if
			// the given permission was authorized on the resource at the matching
			// index.
			authzResults, err := authclient.BatchAuthorize("read", resources)

			// If BatchAuthz failed, we need to return an error to every caller of the
			// loader.
			if err != nil {
				for i := range results {
					results[i] = &dataloader.Result[models.Agency]{
						Error: err,
					}
				}
			}

			// Next we'll perform a batched select query using the ID's that this user
			// did have authz for. But before we do that, we need a way to match a
			// particular ID to its index in the keys array, otherwise the dataloader
			// won't return the correct data to the correct caller of `.Load`.
			var (
				authorizedIds     []string
				authorizedIndexes []int
			)

			// We'll loop through the results of the authz check and assign a zero
			// value to any records the user did not have authz to read, and add an
			// entry to our authorizedResultMap to have the indexes of authorized
			// results handy for the next couple steps.
			for i, authzResult := range authzResults {
				// For the index i, if the user does not have permission, set the Result
				// to a zero result.
				if !authzResult {
					results[i] = &dataloader.Result[models.Agency]{}
				} else {
					id := keys[i]
					authorizedIndexes = append(authorizedIndexes, i)
					authorizedIds = append(authorizedIds, id)
				}
			}

			// If the user isn't authorized to read any of these agencies skip running
			// the query.
			if len(authorizedIds) == 0 {
				return results
			}

			query, args, err := sqlx.In(
				`SELECT id, name, status, created, created_by, modified, modified_by
					 FROM agencies
					 WHERE id IN (?)
					 `,
				authorizedIds)

			// If we failed to generate the query we need to add errors to the
			// dataloader results. However, we need to ensure we only add errors for
			// the items in the result set that would have been in the query (if a
			// user wasn't authorized to read on a particular ID they shouldn't get
			// a SQL error, they should get no result).
			if err != nil {
				for _, index := range authorizedIndexes {
					results[index] = &dataloader.Result[models.Agency]{
						Error: err,
					}
				}
			}

			query = db.Rebind(query)

			rows, err := db.QueryxContext(ctx, query, args...)

			// Like above, if we failed to execute the query we need to add errors to
			// the dataloader results. But we need to only add an error to the result
			// that would have been from an authorized read.
			if err != nil {
				for _, index := range authorizedIndexes {
					results[index] = &dataloader.Result[models.Agency]{
						Error: err,
					}
				}
			}

			// Now we get into really tricky error handling. We'll loop through the
			// rows returned from the query, and attempt to scan each row into a
			// models.Agency struct. If that scan fails, we need to add an error to
			// dataloader results. But because the scan failed, we won't have an ID.
			// Instead we'll use a counter, and find the value of authorizedIndexes at
			// the index of the counter. That will tell us which source key this row
			// read was for.
			rowCount := 0
			for rows.Next() {
				resultIndex := authorizedIndexes[rowCount]
				var agency models.Agency
				if err := rows.StructScan(&agency); err != nil {
					results[resultIndex] = &dataloader.Result[models.Agency]{
						Error: err,
					}
				} else {
					results[resultIndex] = &dataloader.Result[models.Agency]{
						Data: agency,
					}
				}
				rowCount++
			}

			return results
		})
}

type AgencyService struct {
	ctx                    context.Context
	authclient             *authz.Client
	db                     *sqlx.DB
	logger                 *slog.Logger
	validate               *validator.Validate
	listAgenciesDataLoader *dataloader.Loader[AgenciesPagination, []models.Agency]
	readAgencyDataLoader   *dataloader.Loader[string, models.Agency]
}

func NewAgencyService(
	ctx context.Context,
	authz *authz.Client,
	db *sqlx.DB,
	logger *slog.Logger,
	validate *validator.Validate,
) *AgencyService {
	return &AgencyService{
		ctx:                    ctx,
		authclient:             authz,
		db:                     db,
		logger:                 logger,
		validate:               validate,
		listAgenciesDataLoader: newListAgenciesDataLoader(authz, db),
		readAgencyDataLoader:   newReadAgencyDataLoader(authz, db),
	}
}

func (a *AgencyService) List(pagination AgenciesPagination) ([]models.Agency, error) {
	return a.listAgenciesDataLoader.Load(a.ctx, pagination)()
}

func (a *AgencyService) Read(id string) (models.Agency, error) {
	return a.readAgencyDataLoader.Load(a.ctx, id)()
}

func (a *AgencyService) Create(name, userId string) (models.Agency, error) {
	var agency models.Agency

	check, err := a.authclient.Authorize(
		authz.PermissionCreateAgency,
		authz.Resource{Type: "platform", ID: "platform"})

	if err != nil {
		return agency, err
	}
	if !check {
		return agency, authz.NewAuthzError(
			authz.PermissionCreateAgency,
			authz.Resource{Type: "platform", ID: "platform"})
	}

	tx, err := a.db.BeginTxx(a.ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return agency, err
	}

	if err := tx.QueryRowxContext(
		a.ctx,
		`INSERT INTO agencies (name, status, created_by, modified_by)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, name, status, created, created_by, modified, modified_by;`,
		name,
		models.AgencyStatusPending,
		userId,
		userId,
	).StructScan(&agency); err != nil {
		return agency, err
	}

	if err = a.authclient.WritePermission(
		"platform",
		authz.Resource{Type: "agency", ID: agency.ID},
		authz.Resource{Type: "platform", ID: "platform"}); err != nil {
		if txerr := tx.Rollback(); txerr != nil {
			return agency, txerr
		}

		return agency, err
	}

	if err := tx.Commit(); err != nil {
		return agency, err
	}

	return agency, nil
}
