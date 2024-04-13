package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/graph-gophers/dataloader/v7"
	"github.com/jmoiron/sqlx"
	"github.com/jsmithdenverdev/pager/authz"
	"github.com/jsmithdenverdev/pager/models"
)

// PageOrder  represents the sort order in a request to list pages.
type PageOrder int

// PageDeliveryOrder represents the sort order in a request to list page
// deliveries.
type PageDeliveryOrder int

const (
	PageOrderCreatedAsc PageOrder = iota
	PageOrderCreatedDesc
	PageOrderModifiedAsc
	PageOrderModifiedDesc
	PageOrderNameAsc
	PageOrderNameDesc
)

const (
	PageDeliveryOrderCreatedAsc PageDeliveryOrder = iota
	PageDeliveryOrderCreatedDesc
	PageDeliveryOrderModifiedAsc
	PageDeliveryOrderModifiedDesc
	PageDeliveryOrderNameAsc
	PageDeliveryOrderNameDesc
)

var (
	pageOrderNames = [...]string{
		"CREATED_ASC",
		"CREATED_DESC",
		"MODIFIED_ASC",
		"MODIFIED_DESC",
		"NAME_ASC",
		"NAME_DESC",
	}
	pageDeliveryOrderNames = [...]string{
		"CREATED_ASC",
		"CREATED_DESC",
		"MODIFIED_ASC",
		"MODIFIED_DESC",
		"NAME_ASC",
		"NAME_DESC",
	}
	pageOrderMap = map[PageOrder]string{
		PageOrderCreatedAsc:   "created ASC",
		PageOrderCreatedDesc:  "created DESC",
		PageOrderModifiedAsc:  "modified ASC",
		PageOrderModifiedDesc: "modified DESC",
		PageOrderNameAsc:      "name ASC",
		PageOrderNameDesc:     "name DESC",
	}
	pageDeliveryOrderMap = map[PageDeliveryOrder]string{
		PageDeliveryOrderCreatedAsc:   "created ASC",
		PageDeliveryOrderCreatedDesc:  "created DESC",
		PageDeliveryOrderModifiedAsc:  "modified ASC",
		PageDeliveryOrderModifiedDesc: "modified DESC",
		PageDeliveryOrderNameAsc:      "name ASC",
		PageDeliveryOrderNameDesc:     "name DESC",
	}
)

func (order PageDeliveryOrder) String() string {
	return pageDeliveryOrderNames[order]
}

func (order PageOrder) String() string {
	return pageOrderNames[order]
}

type PagePagination struct {
	First  int
	Filter struct {
		AgencyID string
	}
	After string
	Order PageOrder
}

type PageDeliveryPagination struct {
	First  int
	Filter struct {
		PageID   string
		AgencyID string
	}
	After string
	Order PageDeliveryOrder
}

// listPagesDataloader is a request scoped data loader that is used to batch
// page list operations across multiple concurrent resolvers.
func listPagesDataloader(authclient *authz.Client, db *sqlx.DB) *dataloader.Loader[PagePagination, []models.Page] {
	return dataloader.NewBatchedLoader(
		func(ctx context.Context, keys []PagePagination) []*dataloader.Result[[]models.Page] {
			results := make([]*dataloader.Result[[]models.Page], len(keys))
			// Fetch a list of IDs that this user has access to. This data comes from
			// spice db, and we can use it to narrow down our query to the most
			// restrictive set of data for this user.
			ids, err := authclient.List("read", authz.Resource{Type: "page"})

			// If we aren't authorized on any devices return an empty result set
			if len(ids) == 0 {
				for i := range results {
					results[i] = &dataloader.Result[[]models.Page]{}
				}
				return results
			}

			// If List failed, we need to return an error to every caller of the
			// loader.
			if err != nil {
				for i := range results {
					results[i] = &dataloader.Result[[]models.Page]{
						Error: err,
					}
				}

				return results
			}

			for i := range keys {
				var (
					first  = keys[i].First
					order  = pageOrderMap[keys[i].Order]
					after  = keys[i].After
					filter = keys[i].Filter
					query  string
					args   []interface{}
					err    error
				)

				// Create a query using the pagination key. In theory postgres doesn't
				// have an upper limit to the number of values we supply to IN, and the
				// number of agencies one user could possibly belong to is much lower
				// than whatever upper bounds we'd see with postgres.
				// In theory we'd have better performance by performing a single bulk db
				// query but that would prevent each call to load from being able to
				// define its own sort and filters, or we'd have to fetch all the data
				// and do the sorting and filtering in memory which I'm guessing would
				// be slower and more complicated than allowing postgres to do that.
				// Filter on UserID
				query =
					`SELECT id, agency_id, content, created, created_by, modified, modified_by
					 FROM pages
					`

				// Where
				query += "WHERE id IN (:ids)\n"

				if filter.AgencyID != "" {
					query += "AND agency_id = :agencyId\n"
				}

				// After
				if after != "" {
					query += "AND id > :after\n"
				}
				// Ordering
				query += fmt.Sprintf("ORDER BY %s\n", order)
				query += "LIMIT :limit"

				// Fill in parameterized portions of the query
				query, args, err = sqlx.Named(query,
					map[string]interface{}{
						"ids":      ids,
						"agencyId": filter.AgencyID,
						"after":    after,
						"limit":    first,
					})

				// If we failed to create the query, attach an error to the dataloader
				// result for this index. Continue the loop to process the next key in
				// the batch.
				if err != nil {
					results[i] = &dataloader.Result[[]models.Page]{
						Error: err,
					}
					continue
				}

				// Fill the IN clause in the parameterized query
				query, args, err = sqlx.In(query, args...)
				if err != nil {
					results[i] = &dataloader.Result[[]models.Page]{
						Error: err,
					}
					continue
				}

				query = db.Rebind(query)

				// Execute the query
				rows, err := db.QueryxContext(
					ctx,
					query,
					args...,
				)

				// If we failed to execute the query, attach an error to the dataloader
				// result for this index. Continue the loop to process the next key in
				// the batch.
				if err != nil {
					results[i] = &dataloader.Result[[]models.Page]{
						Error: err,
					}
					continue
				}

				// Begin looping through the rows returned in the query. We'll map each
				// row into a `models.Device`. If mapping the row fails, we close the
				// reader to and attach an error to the dataloader result for this
				// index. We break out of the inner for loop to prevent additional calls
				// to the closed reader.
				var pages []models.Page
				for rows.Next() {
					var p models.Page
					if err := rows.StructScan(&p); err != nil {
						results[i] = &dataloader.Result[[]models.Page]{
							Error: err,
						}
						// As we continue operations we need to check for errors and assign
						// them to the dataloader result at for the current index. This will
						// overwrite the result, so we'll only have the most recent error
						// but its enough for us to know where in the stack we failed, and
						// work up from there.
						if err := rows.Close(); err != nil {
							results[i] = &dataloader.Result[[]models.Page]{
								Error: err,
							}
						}
						// Here we break instead of continue. We've closed the db reader
						// and consider the results of this query set a failure.
						break
					}
					pages = append(pages, p)
				}

				// Once we've mapped each agency row into a `models.Device`, we'll add
				// the array of models to the datalaoder result for this index.
				results[i] = &dataloader.Result[[]models.Page]{
					Data: pages,
				}
			}
			return results
		})
}

// readPageDataloader is a request scoped data loader that is used to batch
// device read operations across multiple concurrent resolvers.
func readPageDataloader(authclient *authz.Client, db *sqlx.DB) *dataloader.Loader[string, models.Page] {
	return dataloader.NewBatchedLoader(
		func(ctx context.Context, keys []string) []*dataloader.Result[models.Page] {
			results := make([]*dataloader.Result[models.Page], len(keys))
			var resources []authz.Resource

			// Build up a collection of resources for a batch authorization check. We
			// do this because this dataloader may be called multiple times to fully
			// resolve a particular query. This allows us to coalesce the full set of
			// IDs from all calls to `.Load` into a single authorization query.
			for _, key := range keys {
				resources = append(resources, authz.Resource{Type: "page", ID: key})
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
					results[i] = &dataloader.Result[models.Page]{
						Error: err,
					}
				}

				return results
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
			// entry to our authorizedIndexes array to have the indexes of authorized
			// results handy for the next couple steps.
			for i, authzResult := range authzResults {
				// For the index i, if the user does not have permission, set the Result
				// to a zero result.
				if authzResult.Error != nil {
					results[i] = &dataloader.Result[models.Page]{
						Error: authzResult.Error,
					}
				}
				if !authzResult.Authorized {
					results[i] = &dataloader.Result[models.Page]{}
				} else {
					id := keys[i]
					authorizedIndexes = append(authorizedIndexes, i)
					authorizedIds = append(authorizedIds, id)
				}
			}

			// If the user isn't authorized to read any of these devices skip running
			// the query.
			if len(authorizedIds) == 0 {
				return results
			}

			// Generate our database query, we aren't worried about sorting here
			// because even though this is batching requests, we need to remember that
			// the caller of this method is intending to get a single response.
			query, args, err := sqlx.In(
				`SELECT id, agency_id, content, created, created_by, modified, modified_by
					 FROM pages
					 WHERE id IN (?)`,
				authorizedIds)

			// If we failed to generate the query we need to add errors to the
			// dataloader results. However, we need to ensure we only add errors for
			// the items in the result set that would have been in the query (if a
			// user wasn't authorized to read on a particular ID they shouldn't get
			// a SQL error, they should get no result).
			if err != nil {
				// This is a little odd looking because we're not actually interested
				// in the current index of the range call we're interested in the value
				// at that position in the array. That value corresponds to an index in
				// the result set that would be an authorized read.
				for _, index := range authorizedIndexes {
					results[index] = &dataloader.Result[models.Page]{
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
				// Like above, this is a little odd looking because we're not actually
				// interested in the current index of the range call we're interested in
				// the value at that position in the array. That value corresponds to an
				// index in the result set that would be an authorized read.
				for _, index := range authorizedIndexes {
					results[index] = &dataloader.Result[models.Page]{
						Error: err,
					}
				}
			}

			// We'll loop through the rows returned from the query, and attempt to
			// scan each row into a models.Device struct. If that scan fails, we need
			// to add an error to dataloader result.
			// Because all of our arrays are ordered the same, we can use the rowCount
			// to get a value from authorizedIndexes. That value is the position of
			// this record in the final results array. If we have an error we'll
			// assign an error result to that position, otherwise we'll assign a data
			// result to that position.
			rowCount := 0
			for rows.Next() {
				resultIndex := authorizedIndexes[rowCount]
				var page models.Page
				if err := rows.StructScan(&page); err != nil {
					results[resultIndex] = &dataloader.Result[models.Page]{
						Error: err,
					}
				} else {
					results[resultIndex] = &dataloader.Result[models.Page]{
						Data: page,
					}
				}
				rowCount++
			}

			return results
		})
}

// listDeliveriesDataloader is a request scoped data loader that is used to
// batch page delivery list operations across multiple concurrent resolvers.
func listDeliveriesDataloader(authclient *authz.Client, db *sqlx.DB) *dataloader.Loader[PageDeliveryPagination, []models.PageDelivery] {
	return dataloader.NewBatchedLoader(
		func(ctx context.Context, keys []PageDeliveryPagination) []*dataloader.Result[[]models.PageDelivery] {
			results := make([]*dataloader.Result[[]models.PageDelivery], len(keys))
			// Fetch a list of IDs that this user has access to. This data comes from
			// spice db, and we can use it to narrow down our query to the most
			// restrictive set of data for this user.
			ids, err := authclient.List("read", authz.Resource{Type: "page_delivery"})

			// If we aren't authorized on any devices return an empty result set
			if len(ids) == 0 {
				for i := range results {
					results[i] = &dataloader.Result[[]models.PageDelivery]{}
				}
				return results
			}

			// If List failed, we need to return an error to every caller of the
			// loader.
			if err != nil {
				for i := range results {
					results[i] = &dataloader.Result[[]models.PageDelivery]{
						Error: err,
					}
				}

				return results
			}

			for i := range keys {
				var (
					first  = keys[i].First
					order  = pageDeliveryOrderMap[keys[i].Order]
					after  = keys[i].After
					filter = keys[i].Filter
					query  string
					args   []interface{}
					err    error
				)

				// Create a query using the pagination key. In theory postgres doesn't
				// have an upper limit to the number of values we supply to IN, and the
				// number of agencies one user could possibly belong to is much lower
				// than whatever upper bounds we'd see with postgres.
				// In theory we'd have better performance by performing a single bulk db
				// query but that would prevent each call to load from being able to
				// define its own sort and filters, or we'd have to fetch all the data
				// and do the sorting and filtering in memory which I'm guessing would
				// be slower and more complicated than allowing postgres to do that.
				// Filter on UserID
				query =
					`SELECT pd.id, pd.agency_id, pd.content, pd.created, pd.created_by, pd.modified, pd.modified_by
					 FROM page_deliveries pd
					`

				// Joints
				if filter.AgencyID != "" {
					query += "JOIN pages p on p.id = pd.page_id\n"
				}

				// Where
				query += "WHERE id IN (:ids)\n"

				// Filters
				if filter.PageID != "" {
					query += "AND pd.id = :pageId\n"
				}

				if filter.AgencyID != "" {
					query += "AND p.agency_id = :agencyId\n"
				}

				// After
				if after != "" {
					query += "AND id > :after\n"
				}
				// Ordering
				query += fmt.Sprintf("ORDER BY %s\n", order)
				query += "LIMIT :limit"

				// Fill in parameterized portions of the query
				query, args, err = sqlx.Named(query,
					map[string]interface{}{
						"ids":      ids,
						"pageId":   filter.PageID,
						"agencyId": filter.AgencyID,
						"after":    after,
						"limit":    first,
					})

				// If we failed to create the query, attach an error to the dataloader
				// result for this index. Continue the loop to process the next key in
				// the batch.
				if err != nil {
					results[i] = &dataloader.Result[[]models.PageDelivery]{
						Error: err,
					}
					continue
				}

				// Fill the IN clause in the parameterized query
				query, args, err = sqlx.In(query, args...)
				if err != nil {
					results[i] = &dataloader.Result[[]models.PageDelivery]{
						Error: err,
					}
					continue
				}

				query = db.Rebind(query)

				// Execute the query
				rows, err := db.QueryxContext(
					ctx,
					query,
					args...,
				)

				// If we failed to execute the query, attach an error to the dataloader
				// result for this index. Continue the loop to process the next key in
				// the batch.
				if err != nil {
					results[i] = &dataloader.Result[[]models.PageDelivery]{
						Error: err,
					}
					continue
				}

				// Begin looping through the rows returned in the query. We'll map each
				// row into a `models.Device`. If mapping the row fails, we close the
				// reader to and attach an error to the dataloader result for this
				// index. We break out of the inner for loop to prevent additional calls
				// to the closed reader.
				var deliveries []models.PageDelivery
				for rows.Next() {
					var d models.PageDelivery
					if err := rows.StructScan(&d); err != nil {
						results[i] = &dataloader.Result[[]models.PageDelivery]{
							Error: err,
						}
						// As we continue operations we need to check for errors and assign
						// them to the dataloader result at for the current index. This will
						// overwrite the result, so we'll only have the most recent error
						// but its enough for us to know where in the stack we failed, and
						// work up from there.
						if err := rows.Close(); err != nil {
							results[i] = &dataloader.Result[[]models.PageDelivery]{
								Error: err,
							}
						}
						// Here we break instead of continue. We've closed the db reader
						// and consider the results of this query set a failure.
						break
					}
					deliveries = append(deliveries, d)
				}

				// Once we've mapped each agency row into a `models.Device`, we'll add
				// the array of models to the datalaoder result for this index.
				results[i] = &dataloader.Result[[]models.PageDelivery]{
					Data: deliveries,
				}
			}
			return results
		})
}

// readDeliveryDataloader is a request scoped data loader that is used to batch
// device read operations across multiple concurrent resolvers.
func readDeliveryDataloader(authclient *authz.Client, db *sqlx.DB) *dataloader.Loader[string, models.PageDelivery] {
	return dataloader.NewBatchedLoader(
		func(ctx context.Context, keys []string) []*dataloader.Result[models.PageDelivery] {
			results := make([]*dataloader.Result[models.PageDelivery], len(keys))
			var resources []authz.Resource

			// Build up a collection of resources for a batch authorization check. We
			// do this because this dataloader may be called multiple times to fully
			// resolve a particular query. This allows us to coalesce the full set of
			// IDs from all calls to `.Load` into a single authorization query.
			for _, key := range keys {
				resources = append(resources, authz.Resource{Type: "page_delivery", ID: key})
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
					results[i] = &dataloader.Result[models.PageDelivery]{
						Error: err,
					}
				}

				return results
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
			// entry to our authorizedIndexes array to have the indexes of authorized
			// results handy for the next couple steps.
			for i, authzResult := range authzResults {
				// For the index i, if the user does not have permission, set the Result
				// to a zero result.
				if authzResult.Error != nil {
					results[i] = &dataloader.Result[models.PageDelivery]{
						Error: authzResult.Error,
					}
				}
				if !authzResult.Authorized {
					results[i] = &dataloader.Result[models.PageDelivery]{}
				} else {
					id := keys[i]
					authorizedIndexes = append(authorizedIndexes, i)
					authorizedIds = append(authorizedIds, id)
				}
			}

			// If the user isn't authorized to read any of these devices skip running
			// the query.
			if len(authorizedIds) == 0 {
				return results
			}

			// Generate our database query, we aren't worried about sorting here
			// because even though this is batching requests, we need to remember that
			// the caller of this method is intending to get a single response.
			query, args, err := sqlx.In(
				`SELECT id, page_id, device_id, status, created, created_by, modified, modified_by
					 FROM page_deliveries
					 WHERE id IN (?)`,
				authorizedIds)

			// If we failed to generate the query we need to add errors to the
			// dataloader results. However, we need to ensure we only add errors for
			// the items in the result set that would have been in the query (if a
			// user wasn't authorized to read on a particular ID they shouldn't get
			// a SQL error, they should get no result).
			if err != nil {
				// This is a little odd looking because we're not actually interested
				// in the current index of the range call we're interested in the value
				// at that position in the array. That value corresponds to an index in
				// the result set that would be an authorized read.
				for _, index := range authorizedIndexes {
					results[index] = &dataloader.Result[models.PageDelivery]{
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
				// Like above, this is a little odd looking because we're not actually
				// interested in the current index of the range call we're interested in
				// the value at that position in the array. That value corresponds to an
				// index in the result set that would be an authorized read.
				for _, index := range authorizedIndexes {
					results[index] = &dataloader.Result[models.PageDelivery]{
						Error: err,
					}
				}
			}

			// We'll loop through the rows returned from the query, and attempt to
			// scan each row into a models.Device struct. If that scan fails, we need
			// to add an error to dataloader result.
			// Because all of our arrays are ordered the same, we can use the rowCount
			// to get a value from authorizedIndexes. That value is the position of
			// this record in the final results array. If we have an error we'll
			// assign an error result to that position, otherwise we'll assign a data
			// result to that position.
			rowCount := 0
			for rows.Next() {
				resultIndex := authorizedIndexes[rowCount]
				var delivery models.PageDelivery
				if err := rows.StructScan(&delivery); err != nil {
					results[resultIndex] = &dataloader.Result[models.PageDelivery]{
						Error: err,
					}
				} else {
					results[resultIndex] = &dataloader.Result[models.PageDelivery]{
						Data: delivery,
					}
				}
				rowCount++
			}

			return results
		})
}

// DeviceService exposes all operations that can be performed on or for devices.
type PageService struct {
	ctx                      context.Context
	user                     string
	authclient               *authz.Client
	db                       *sqlx.DB
	logger                   *slog.Logger
	listPagesDataloader      *dataloader.Loader[PagePagination, []models.Page]
	readPageDataloader       *dataloader.Loader[string, models.Page]
	listDeliveriesDataloader *dataloader.Loader[PageDeliveryPagination, []models.PageDelivery]
	readDeliveryDataloader   *dataloader.Loader[string, models.PageDelivery]
}

// NewDeviceService creates a new DeviceService. A pointer to the service is
// returned.
func NewPageService(
	ctx context.Context,
	user string,
	authz *authz.Client,
	db *sqlx.DB,
	logger *slog.Logger,
) *PageService {
	return &PageService{
		ctx:                      ctx,
		user:                     user,
		authclient:               authz,
		db:                       db,
		logger:                   logger,
		listPagesDataloader:      listPagesDataloader(authz, db),
		readPageDataloader:       readPageDataloader(authz, db),
		listDeliveriesDataloader: listDeliveriesDataloader(authz, db),
		readDeliveryDataloader:   readDeliveryDataloader(authz, db),
	}
}

func (service *PageService) CreatePage(agencyId string, content string, deliver bool) (models.Page, error) {
	var page models.Page
	return page, nil
}

func (service *PageService) ReadPage(id string) (models.Page, error) {
	return service.readPageDataloader.Load(service.ctx, id)()
}

func (service *PageService) UpdatePage(content string) (models.Page, error) {
	var page models.Page
	return page, nil
}

func (service *PageService) ListPages(pagination PagePagination) ([]models.Page, error) {
	return service.listPagesDataloader.Load(service.ctx, pagination)()
}

func (service *PageService) DeletePage(id string) error {
	return nil
}

func (service *PageService) DeliverPage(pageId string) error {
	return nil
}

func (service *PageService) ReadDelivery(id string) (models.PageDelivery, error) {
	return service.readDeliveryDataloader.Load(service.ctx, id)()
}

func (service *PageService) ListDeliveries(pagination PageDeliveryPagination) ([]models.PageDelivery, error) {
	return service.listDeliveriesDataloader.Load(service.ctx, pagination)()
}
