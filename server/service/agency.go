package service

import (
	"context"
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

func newListAgenciesDataLoader(authz authz.Client, db *sqlx.DB) *dataloader.Loader[AgenciesPagination, []*models.Agency] {
	return dataloader.NewBatchedLoader(
		func(ctx context.Context, paginations []AgenciesPagination) []*dataloader.Result[[]*models.Agency] {
			results := make([]*dataloader.Result[[]*models.Agency], len(paginations))
			// Give me the list of agencies this user has the read permission on
			ids, err := authz.List("read", "agency")
			// We're not actually using pagination settings yet, I just want to get
			// this working through a loader to start.
			for i, _ := range paginations {
				if err != nil {
					results[i] = &dataloader.Result[[]*models.Agency]{
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
					results[i] = &dataloader.Result[[]*models.Agency]{
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
					results[i] = &dataloader.Result[[]*models.Agency]{
						Error: err,
					}
					continue
				}

				var agencies []*models.Agency
				for rows.Next() {
					var a *models.Agency
					if err := rows.StructScan(a); err != nil {
						results[i] = &dataloader.Result[[]*models.Agency]{
							Error: err,
						}
						if err := rows.Close(); err != nil {
							results[i] = &dataloader.Result[[]*models.Agency]{
								Error: err,
							}
						}
						// Here we break instead of continue. This closes the db connection
						// and considers this entire query set a failure.
						break
					}
					agencies = append(agencies, a)
				}

				results[i] = &dataloader.Result[[]*models.Agency]{
					Data: agencies,
				}
			}
			return results
		})
}

type AgencyService struct {
	ctx                    context.Context
	authz                  authz.Client
	db                     *sqlx.DB
	logger                 *slog.Logger
	validate               *validator.Validate
	listAgenciesDataLoader *dataloader.Loader[AgenciesPagination, []*models.Agency]
}

func NewAgencyService(
	ctx context.Context,
	authz authz.Client,
	db *sqlx.DB,
	logger *slog.Logger,
	validate *validator.Validate,
) *AgencyService {
	return &AgencyService{
		ctx:                    ctx,
		authz:                  authz,
		db:                     db,
		logger:                 logger,
		validate:               validate,
		listAgenciesDataLoader: newListAgenciesDataLoader(authz, db),
	}
}

func (a *AgencyService) List(pagination AgenciesPagination) ([]*models.Agency, error) {
	return a.listAgenciesDataLoader.Load(a.ctx, pagination)()
}
