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

// UsersOrder  represents the sort order in a request to list users for an
// agency.
type UsersOrder int

const (
	UsersOrderCreatedAsc UsersOrder = iota
	UsersOrderCreatedDesc
	UsersOrderModifiedAsc
	UsersOrderModifiedDesc
	UsersOrderNameAsc
	UsersOrderNameDesc
)

var (
	usersOrderNames = [...]string{
		"CREATED_ASC",
		"CREATED_DESC",
		"MODIFIED_ASC",
		"MODIFIED_DESC",
		"NAME_ASC",
		"NAME_DESC",
	}
	userOrderMap = map[UsersOrder]string{
		UsersOrderCreatedAsc:   "created ASC",
		UsersOrderCreatedDesc:  "created DESC",
		UsersOrderModifiedAsc:  "modified ASC",
		UsersOrderModifiedDesc: "modified DESC",
		UsersOrderNameAsc:      "name ASC",
		UsersOrderNameDesc:     "name DESC",
	}
)

func (order UsersOrder) String() string {
	return usersOrderNames[order]
}

type UsersPagination struct {
	First  int
	After  string
	Order  UsersOrder
	Filter struct {
		AgencyID string
	}
}

type UserService struct {
	ctx                 context.Context
	user                string
	authclient          *authz.Client
	db                  *sqlx.DB
	listUsersDataloader *dataloader.Loader[UsersPagination, []models.User]
	logger              *slog.Logger
}

// listUsersDataloader is a request scoped data loader that is used to batch
// user list operations across multiple concurrent resolvers.
func listUsersDataloader(authclient *authz.Client, db *sqlx.DB) *dataloader.Loader[UsersPagination, []models.User] {
	return dataloader.NewBatchedLoader(
		func(ctx context.Context, keys []UsersPagination) []*dataloader.Result[[]models.User] {
			results := make([]*dataloader.Result[[]models.User], len(keys))
			// Fetch a list of IDs that this user has access to. This data comes from
			// spice db, and we can use it to narrow down our query to the most
			// restrictive set of data for this user.
			ids, err := authclient.List("read", authz.Resource{Type: "user"})

			// If we aren't authorized on any agencies return an empty result set
			if len(ids) == 0 {
				for i := range results {
					results[i] = &dataloader.Result[[]models.User]{}
				}
				return results
			}

			// If List failed, we need to return an error to every caller of the
			// loader.
			if err != nil {
				for i := range results {
					results[i] = &dataloader.Result[[]models.User]{
						Error: err,
					}
				}

				return results
			}

			for i := range keys {
				var (
					first  = keys[i].First
					order  = userOrderMap[keys[i].Order]
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
				query =
					`SELECT id, idp_id, email, status, created, created_by, modified, modified_by
				 FROM users
				`

					// Where
				query += "WHERE idp_id IN (:ids)\n"

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
					results[i] = &dataloader.Result[[]models.User]{
						Error: err,
					}
					continue
				}

				// Fill the IN clause in the parameterized query
				query, args, err = sqlx.In(query, args...)
				if err != nil {
					results[i] = &dataloader.Result[[]models.User]{
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
					results[i] = &dataloader.Result[[]models.User]{
						Error: err,
					}
					continue
				}

				// Begin looping through the rows returned in the query. We'll map each
				// row into a `models.Device`. If mapping the row fails, we close the
				// reader to and attach an error to the dataloader result for this
				// index. We break out of the inner for loop to prevent additional calls
				// to the closed reader.
				var users []models.User
				for rows.Next() {
					var u models.User
					if err := rows.StructScan(&u); err != nil {
						results[i] = &dataloader.Result[[]models.User]{
							Error: err,
						}
						// As we continue operations we need to check for errors and assign
						// them to the dataloader result at for the current index. This will
						// overwrite the result, so we'll only have the most recent error
						// but its enough for us to know where in the stack we failed, and
						// work up from there.
						if err := rows.Close(); err != nil {
							results[i] = &dataloader.Result[[]models.User]{
								Error: err,
							}
						}
						// Here we break instead of continue. We've closed the db reader
						// and consider the results of this query set a failure.
						break
					}
					users = append(users, u)
				}

				// Once we've mapped each agency row into a `models.Device`, we'll add
				// the array of models to the datalaoder result for this index.
				results[i] = &dataloader.Result[[]models.User]{
					Data: users,
				}
			}
			return results
		})
}

func NewUserService(ctx context.Context, user string, authz *authz.Client, db *sqlx.DB, logger *slog.Logger) *UserService {
	return &UserService{
		ctx:                 ctx,
		user:                user,
		authclient:          authz,
		db:                  db,
		listUsersDataloader: listUsersDataloader(authz, db),
		logger:              logger,
	}
}

func (service *UserService) Info() (models.User, error) {
	var user models.User
	if err := service.db.QueryRowxContext(
		service.ctx,
		`SELECT id, email, idp_id, status, created, created_by, modified, modified_by
		 FROM users
		 WHERE idp_id = $1`,
		service.user,
	).StructScan(&user); err != nil {
		return user, err
	}

	return user, nil
}

func (service *UserService) Roles() ([]models.UserRole, error) {
	// TODO: Put this in a dataloader
	var roles []models.UserRole

	rows, err := service.db.QueryxContext(
		service.ctx,
		`SELECT role, user_id, agency_id
		 FROM user_roles
		 INNER JOIN users on users.id = user_roles.user_id
		 WHERE idp_id = $1`,
		service.user,
	)

	if err != nil {
		return roles, err
	}

	for rows.Next() {
		var role models.UserRole
		if err := rows.StructScan(&role); err != nil {
			return roles, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (service *UserService) ListUsers(pagination UsersPagination) ([]models.User, error) {
	return service.listUsersDataloader.Load(service.ctx, pagination)()
}
