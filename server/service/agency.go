package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/auth0/go-auth0/management"
	"github.com/graph-gophers/dataloader/v7"
	"github.com/jmoiron/sqlx"
	"github.com/jsmithdenverdev/pager/authz"
	"github.com/jsmithdenverdev/pager/models"
	"github.com/jsmithdenverdev/pager/pubsub"
)

// AgenciesOrder represents the sort order in a request to list agencies.
type AgenciesOrder int

const (
	AgenciesOrderCreatedAsc AgenciesOrder = iota
	AgenciesOrderCreatedDesc
	AgenciesOrderModifiedAsc
	AgenciesOrderModifiedDesc
	AgenciesOrderNameAsc
	AgenciesOrderNameDesc
)

var (
	agenciesOrderNames = [...]string{
		"CREATED_ASC",
		"CREATED_DESC",
		"MODIFIED_ASC",
		"MODIFIED_DESC",
		"NAME_ASC",
		"NAME_DESC",
	}
	agencyOrderMap = map[AgenciesOrder]string{
		AgenciesOrderCreatedAsc:   "created ASC",
		AgenciesOrderCreatedDesc:  "created DESC",
		AgenciesOrderModifiedAsc:  "modified ASC",
		AgenciesOrderModifiedDesc: "modified DESC",
		AgenciesOrderNameAsc:      "name ASC",
		AgenciesOrderNameDesc:     "name DESC",
	}
)

func (order AgenciesOrder) String() string {
	return agenciesOrderNames[order]
}

type AgenciesPagination struct {
	First int
	After string
	Order AgenciesOrder
}

// listAgenciesDataloader is a request scoped data loader that is used to batch
// agency list operations across multiple concurrent resolvers.
func listAgenciesDataloader(authclient *authz.Client, db *sqlx.DB) *dataloader.Loader[AgenciesPagination, []models.Agency] {
	return dataloader.NewBatchedLoader(
		func(ctx context.Context, keys []AgenciesPagination) []*dataloader.Result[[]models.Agency] {
			results := make([]*dataloader.Result[[]models.Agency], len(keys))
			// Fetch a list of IDs that this user has access to. This data comes from
			// spice db, and we can use it to narrow down our query to the most
			// restrictive set of data for this user.
			ids, err := authclient.List("read", authz.Resource{Type: "agency"})

			// If we aren't authorized on any agencies return an empty result set
			if len(ids) == 0 {
				for i := range results {
					results[i] = &dataloader.Result[[]models.Agency]{}
				}
				return results
			}

			// If List failed, we need to return an error to every caller of the
			// loader.
			if err != nil {
				for i := range results {
					results[i] = &dataloader.Result[[]models.Agency]{
						Error: err,
					}
				}

				return results
			}

			for i := range keys {
				var (
					first = keys[i].First
					order = agencyOrderMap[keys[i].Order]
					after = keys[i].After
					query string
					args  []interface{}
					err   error
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
				if after == "" {
					// Create the initial query, ORDER BY clauses can't be parameterized so
					// we must use regular ol' Sprintf statements to parameterize the ORDER
					// BY clause.
					query = fmt.Sprintf(
						`SELECT id, name, status, created, created_by, modified, modified_by
					FROM agencies
					WHERE id in (:ids)
					ORDER BY %s
					LIMIT :limit`,
						order)

					// Fill in parameterized portions of the query
					query, args, err = sqlx.Named(query,
						map[string]interface{}{
							"ids":   ids,
							"limit": first,
						})

					// If we failed to create the query, attach an error to the dataloader
					// result for this index. Continue the loop to process the next key in
					// the batch.
					if err != nil {
						results[i] = &dataloader.Result[[]models.Agency]{
							Error: err,
						}
						continue
					}
				} else {
					query = fmt.Sprintf(
						`SELECT id, name, status, created, created_by, modified, modified_by
					FROM agencies
					WHERE id in (:ids)
					AND id > :after
					ORDER BY %s
					LIMIT :limit`,
						order)

					// Fill in parameterized portions of the query
					query, args, err = sqlx.Named(query,
						map[string]interface{}{
							"ids":   ids,
							"after": after,
							"limit": first,
						})

					// If we failed to create the query, attach an error to the dataloader
					// result for this index. Continue the loop to process the next key in
					// the batch.
					if err != nil {
						results[i] = &dataloader.Result[[]models.Agency]{
							Error: err,
						}
						continue
					}
				}

				// Fill the IN clause in the parameterized query
				query, args, err = sqlx.In(query, args...)
				if err != nil {
					results[i] = &dataloader.Result[[]models.Agency]{
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
					results[i] = &dataloader.Result[[]models.Agency]{
						Error: err,
					}
					continue
				}

				// Begin looping through the rows returned in the query. We'll map each
				// row into a `models.Agency`. If mapping the row fails, we close the
				// reader to and attach an error to the dataloader result for this
				// index. We break out of the inner for loop to prevent additional calls
				// to the closed reader.
				var agencies []models.Agency
				for rows.Next() {
					var a models.Agency
					if err := rows.StructScan(&a); err != nil {
						results[i] = &dataloader.Result[[]models.Agency]{
							Error: err,
						}
						// As we continue operations we need to check for errors and assign
						// them to the dataloader result at for the current index. This will
						// overwrite the result, so we'll only have the most recent error
						// but its enough for us to know where in the stack we failed, and
						// work up from there.
						if err := rows.Close(); err != nil {
							results[i] = &dataloader.Result[[]models.Agency]{
								Error: err,
							}
						}
						// Here we break instead of continue. We've closed the db reader
						// and consider the results of this query set a failure.
						break
					}
					agencies = append(agencies, a)
				}

				// Once we've mapped each agency row into a `models.Agency`, we'll add
				// the array of models to the datalaoder result for this index.
				results[i] = &dataloader.Result[[]models.Agency]{
					Data: agencies,
				}
			}
			return results
		})
}

// readAgencyDataloader is a request scoped data loader that is used to batch
// agency read operations across multiple concurrent resolvers.
func readAgencyDataloader(authclient *authz.Client, db *sqlx.DB) *dataloader.Loader[string, models.Agency] {
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
					results[i] = &dataloader.Result[models.Agency]{
						Error: authzResult.Error,
					}
				}
				if !authzResult.Authorized {
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

			// Generate our database query, we aren't worried about sorting here
			// because even though this is batching requests, we need to rememebr that
			// the caller of this method is intending to get a single response.
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
				// This is a little odd looking because we're not actually interested
				// in the current index of the range call we're interested in the value
				// at that position in the array. That value corresponds to an index in
				// the result set that would be an authorized read.
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
				// Like above, this is a little odd looking because we're not actually
				// interested in the current index of the range call we're interested in
				// the value at that position in the array. That value corresponds to an
				// index in the result set that would be an authorized read.
				for _, index := range authorizedIndexes {
					results[index] = &dataloader.Result[models.Agency]{
						Error: err,
					}
				}
			}

			// We'll loop through the rows returned from the query, and attempt to
			// scan each row into a models.Agency struct. If that scan fails, we need
			// to add an error to dataloader result.
			// Because all of our arrays are ordered the same, we can use the rowCount
			// to get a value from authorizedIndexes. That value is the position of
			// this record in the final results array. If we have an error we'll
			// assign an error result to that position, otherwise we'll assign a data
			// result to that position.
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

// AgencyService exposes all operations that can be performed on agencies.
type AgencyService struct {
	ctx                    context.Context
	user                   string
	authclient             *authz.Client
	db                     *sqlx.DB
	auth0                  *management.Management
	pubsub                 *pubsub.Client
	logger                 *slog.Logger
	listAgenciesDataLoader *dataloader.Loader[AgenciesPagination, []models.Agency]
	readAgencyDataLoader   *dataloader.Loader[string, models.Agency]
}

// NewAgencyService creates a new AgencyService. A pointer to the service is
// returned.
func NewAgencyService(
	ctx context.Context,
	user string,
	authz *authz.Client,
	db *sqlx.DB,
	auth0 *management.Management,
	pubsub *pubsub.Client,
	logger *slog.Logger,
) *AgencyService {
	return &AgencyService{
		ctx:                    ctx,
		user:                   user,
		authclient:             authz,
		db:                     db,
		auth0:                  auth0,
		pubsub:                 pubsub,
		logger:                 logger,
		listAgenciesDataLoader: listAgenciesDataloader(authz, db),
		readAgencyDataLoader:   readAgencyDataloader(authz, db),
	}
}

// ListAgencies allows a user to list agencies using a set of pagination options.
// Listing is backed by a dataloader making this method safe to use in
// resolvers.
func (service *AgencyService) ListAgencies(pagination AgenciesPagination) ([]models.Agency, error) {
	return service.listAgenciesDataLoader.Load(service.ctx, pagination)()
}

func (service *AgencyService) ReadAgency(id string) (models.Agency, error) {
	return service.readAgencyDataLoader.Load(service.ctx, id)()
}

func (service *AgencyService) CreateAgency(name string) (models.Agency, error) {
	var agency models.Agency

	result := service.authclient.Authorize(
		authz.PermissionCreateAgency,
		authz.Resource{Type: "platform", ID: "platform"})

	if result.Error != nil {
		return agency, result.Error
	}
	if !result.Authorized {
		return agency, authz.NewAuthzError(
			authz.PermissionCreateAgency,
			authz.Resource{Type: "platform", ID: "platform"})
	}

	tx, err := service.db.BeginTxx(service.ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return agency, err
	}

	if err := tx.QueryRowxContext(
		service.ctx,
		`INSERT INTO agencies (name, status, created_by, modified_by)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, name, status, created, created_by, modified, modified_by;`,
		name,
		models.AgencyStatusPending,
		service.user,
		service.user,
	).StructScan(&agency); err != nil {
		return agency, err
	}

	if err = service.authclient.WritePermissions([]authz.Permission{
		{
			Relationship: "platform",
			Resource:     authz.Resource{Type: "agency", ID: agency.ID},
			Subject:      authz.Resource{Type: "platform", ID: "platform"},
		},
	}); err != nil {
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

// InviteUser allows an agency admin to invite a new user to their agency. A
// user record will be created if this is a new user to the platform, or the
// existing user record will be returned after the association is created.
func (service *AgencyService) InviteUser(email string, agencyId string) (models.User, error) {
	var (
		user models.User
	)

	authzResult := service.authclient.Authorize(authz.PermissionInviteUser, authz.Resource{
		Type: "agency",
		ID:   agencyId,
	})

	if authzResult.Error != nil {
		return user, authzResult.Error
	}

	if !authzResult.Authorized {
		return user, authz.NewAuthzError(authz.PermissionInviteUser, authz.Resource{
			Type: "agency",
			ID:   agencyId,
		})
	}

	if err := service.db.QueryRowxContext(
		service.ctx,
		`SELECT id, email, idp_id, status, created, created_by, modified, modified_by
	 	 FROM users
		 WHERE email = $1`,
		email).StructScan(&user); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return user, err
		}
	}

	tx, err := service.db.BeginTxx(service.ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})

	if err != nil {
		return user, err
	}

	// No ID on the user, indicating no record found above. We'll need to create
	// a new User record.
	if user.ID == "" {
		if err := tx.QueryRowxContext(
			service.ctx,
			`INSERT INTO users (email, idp_id, status, created_by, modified_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, email, idp_id, status, created, created_by, modified, modified_by`,
			email,
			"",
			models.UserStatusPending,
			service.user,
			service.user,
		).StructScan(&user); err != nil {
			return user, err
		}
	}

	// No IdpId found on the user. This can happen 1 of two ways.
	// 1. This is a new user record and we just created a new row above.
	// 2. We successfully created all the records, but the ProvisionUserHandler
	//    fails to provision the user.
	// In either case, we write a new provision user message for our handler to
	// consume.
	// In the event the handler failed, we also need to check the messages_dl
	// table and system logs for relevant details.
	if user.IdpID == "" {
		// Create a payload and marshal it into json. The payload column of the
		// messages table is jsonb.
		payload, err := json.Marshal(pubsub.MessageProvisionUser{Email: email})
		if err != nil {
			return user, err
		}

		message := map[string]interface{}{
			"topic":       pubsub.TopicProvisionUser,
			"payload":     payload,
			"created_by":  service.user,
			"modified_by": service.user,
		}

		if _, err := tx.NamedExecContext(
			service.ctx,
			`INSERT INTO messages (topic, payload, created_by, modified_by)
		VALUES (:topic, :payload, :created_by, :modified_by)`,
			message); err != nil {
			return user, err
		}
	}

	// Check for an existing User to Agency association. If this year is not yet
	// associated to the inviting agency, we'll create the association. Otherwise
	// we'll return an error stating that the user is already a member.
	var userAgencyUserId string
	if err := tx.QueryRowxContext(service.ctx,
		`SELECT user_id
	FROM user_agencies
	WHERE user_id = $1
	AND agency_id = $2`,
		user.ID,
		agencyId).Scan(&userAgencyUserId); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return user, err
		}
	}

	if userAgencyUserId == "" {
		// Insert a record into the user_agencies table. If the user is already added
		// return a new error so the caller doesn't keep trying to add the user.
		if _, err := tx.ExecContext(
			service.ctx,
			`INSERT INTO user_agencies (user_id, agency_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING`,
			user.ID,
			agencyId,
		); err != nil {
			return user, err
		}
	} else {
		return user, errors.New("user already member of agency")
	}

	// Commit all of our changes. The write to the messages table will also
	// a function that calls pg_notify using the topic name and message payload.
	if err := tx.Commit(); err != nil {
		return user, err
	}

	return user, nil
}
