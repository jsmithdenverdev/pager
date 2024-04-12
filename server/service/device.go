package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/graph-gophers/dataloader/v7"
	"github.com/jmoiron/sqlx"
	"github.com/jsmithdenverdev/pager/authz"
	"github.com/jsmithdenverdev/pager/models"
)

// DeviceOrder represents the sort order in a request to list devices.
type DeviceOrder int

const (
	DeviceOrderCreatedAsc DeviceOrder = iota
	DeviceOrderCreatedDesc
	DeviceOrderModifiedAsc
	DeviceOrderModifiedDesc
	DeviceOrderNameAsc
	DeviceOrderNameDesc
)

var (
	deviceOrderNames = [...]string{
		"CREATED_ASC",
		"CREATED_DESC",
		"MODIFIED_ASC",
		"MODIFIED_DESC",
		"NAME_ASC",
		"NAME_DESC",
	}
	deviceOrderMap = map[DeviceOrder]string{
		DeviceOrderCreatedAsc:   "created ASC",
		DeviceOrderCreatedDesc:  "created DESC",
		DeviceOrderModifiedAsc:  "modified ASC",
		DeviceOrderModifiedDesc: "modified DESC",
		DeviceOrderNameAsc:      "name ASC",
		DeviceOrderNameDesc:     "name DESC",
	}
)

func (order DeviceOrder) String() string {
	return deviceOrderNames[order]
}

type DevicePagination struct {
	First  int
	Filter struct {
		AgencyID string
		UserID   string
	}
	After string
	Order DeviceOrder
}

// listDevicesDataloader is a request scoped data loader that is used to batch
// agency list operations across multiple concurrent resolvers.
func listDevicesDataloader(authclient *authz.Client, db *sqlx.DB) *dataloader.Loader[DevicePagination, []models.Device] {
	return dataloader.NewBatchedLoader(
		func(ctx context.Context, keys []DevicePagination) []*dataloader.Result[[]models.Device] {
			results := make([]*dataloader.Result[[]models.Device], len(keys))
			// Fetch a list of IDs that this user has access to. This data comes from
			// spice db, and we can use it to narrow down our query to the most
			// restrictive set of data for this user.
			ids, err := authclient.List("read", authz.Resource{Type: "device"})

			// If we aren't authorized on any devices return an empty result set
			if len(ids) == 0 {
				for i := range results {
					results[i] = &dataloader.Result[[]models.Device]{}
				}
				return results
			}

			// If List failed, we need to return an error to every caller of the
			// loader.
			if err != nil {
				for i := range results {
					results[i] = &dataloader.Result[[]models.Device]{
						Error: err,
					}
				}

				return results
			}

			for i := range keys {
				var (
					first  = keys[i].First
					order  = deviceOrderMap[keys[i].Order]
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
					`SELECT d.id, d.name, d.status, d.user_id, d.code, endpoint, d.created, d.created_by, d.modified, d.modified_by
					 FROM devices d
					`

				// Joins
				if filter.AgencyID != "" {
					query += "JOIN agency_devices ad on ad.device_id = d.id\n"
				}

				// Wheres
				query += "WHERE d.id IN (:ids)\n"

				// Filters
				if filter.UserID != "" {
					query += "AND d.user_id = :userId\n"
				}

				if filter.AgencyID != "" {
					query += "AND ad.agency_id = :agencyId\n"
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
						"userId":   filter.UserID,
						"agencyId": filter.AgencyID,
						"after":    after,
						"limit":    first,
					})

				// If we failed to create the query, attach an error to the dataloader
				// result for this index. Continue the loop to process the next key in
				// the batch.
				if err != nil {
					results[i] = &dataloader.Result[[]models.Device]{
						Error: err,
					}
					continue
				}

				// Fill the IN clause in the parameterized query
				query, args, err = sqlx.In(query, args...)
				if err != nil {
					results[i] = &dataloader.Result[[]models.Device]{
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
					results[i] = &dataloader.Result[[]models.Device]{
						Error: err,
					}
					continue
				}

				// Begin looping through the rows returned in the query. We'll map each
				// row into a `models.Device`. If mapping the row fails, we close the
				// reader to and attach an error to the dataloader result for this
				// index. We break out of the inner for loop to prevent additional calls
				// to the closed reader.
				var devices []models.Device
				for rows.Next() {
					var d models.Device
					if err := rows.StructScan(&d); err != nil {
						results[i] = &dataloader.Result[[]models.Device]{
							Error: err,
						}
						// As we continue operations we need to check for errors and assign
						// them to the dataloader result at for the current index. This will
						// overwrite the result, so we'll only have the most recent error
						// but its enough for us to know where in the stack we failed, and
						// work up from there.
						if err := rows.Close(); err != nil {
							results[i] = &dataloader.Result[[]models.Device]{
								Error: err,
							}
						}
						// Here we break instead of continue. We've closed the db reader
						// and consider the results of this query set a failure.
						break
					}
					devices = append(devices, d)
				}

				// Once we've mapped each agency row into a `models.Device`, we'll add
				// the array of models to the datalaoder result for this index.
				results[i] = &dataloader.Result[[]models.Device]{
					Data: devices,
				}
			}
			return results
		})
}

// DeviceService exposes all operations that can be performed on or for devices.
type DeviceService struct {
	ctx                   context.Context
	user                  string
	authclient            *authz.Client
	db                    *sqlx.DB
	logger                *slog.Logger
	listDevicesDataloader *dataloader.Loader[DevicePagination, []models.Device]
}

// NewDeviceService creates a new DeviceService. A pointer to the service is
// returned.
func NewDeviceService(
	ctx context.Context,
	user string,
	authz *authz.Client,
	db *sqlx.DB,
	logger *slog.Logger,
) *DeviceService {
	return &DeviceService{
		ctx:                   ctx,
		user:                  user,
		authclient:            authz,
		db:                    db,
		logger:                logger,
		listDevicesDataloader: listDevicesDataloader(authz, db),
	}
}

// ProvisionDevice provisions a new device to a user. A device belongs to a
// user and can be associated with multiple agencies. A user cannot provision
// their own device. Provisioning must be done by an agency admin.
// Once a device has been provisoned it must be activated by the user. This is
// done by the user logging into the Pager app on their device and entering the
// unique device code linked to this device.
// Agency admins can add activated devices to their agency, and the provisioning
// process adds the device to the given agency automatically.
// In a billing scenario, an agency would be billed for the number of active
// devices associated with it.
func (service *DeviceService) ProvisionDevice(agencyId, ownerId, name string) (models.Device, error) {
	var device models.Device

	result := service.authclient.Authorize(
		authz.PermissionProvisionDevice,
		authz.Resource{Type: "agency", ID: agencyId},
	)

	if result.Error != nil {
		return device, result.Error
	}

	if !result.Authorized {
		return device, authz.NewAuthzError(
			authz.PermissionProvisionDevice,
			authz.Resource{Type: "agency", ID: agencyId},
		)
	}

	tx, err := service.db.BeginTxx(service.ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})

	if err != nil {
		return device, err
	}

	var ownerIdpID string
	if err := tx.QueryRowxContext(service.ctx,
		`SELECT u.idp_id
		 FROM users u
		 INNER JOIN user_agencies ua on ua.user_id = u.id
		 INNER JOIN agencies a on a.id = ua.agency_id
		 WHERE u.id = $1
		 AND a.id = $2`,
		ownerId,
		agencyId,
	).Scan(&ownerIdpID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return device, errors.New("could not find user")
		}
		return device, err
	}

	if err := tx.QueryRowxContext(
		service.ctx,
		`INSERT INTO devices (name, status, user_id, code, created_by, modified_by)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, name, status, code, created, created_by, modified, modified_by;`,
		name,
		models.DeviceStatusPending,
		ownerId,
		generateDeviceCode(10),
		service.user,
		service.user,
	).StructScan(&device); err != nil {
		return device, err
	}

	if _, err := tx.ExecContext(
		service.ctx,
		`INSERT INTO agency_devices (agency_id, device_id)
		 VALUES ($1, $2);`,
		agencyId,
		device.ID,
	); err != nil {
		return device, err
	}

	if err := service.authclient.WritePermission(
		"agency",
		authz.Resource{Type: "device", ID: device.ID},
		authz.Resource{Type: "agency", ID: agencyId},
	); err != nil {
		if txerr := tx.Rollback(); txerr != nil {
			return device, txerr
		}
		return device, err
	}

	if err := service.authclient.WritePermission(
		"owner",
		authz.Resource{Type: "device", ID: device.ID},
		authz.Resource{Type: "user", ID: ownerIdpID},
	); err != nil {
		if txerr := tx.Rollback(); txerr != nil {
			return device, txerr
		}
		return device, err
	}

	// If we fail to commit the transaction, we'll still have relations published
	// to spicedb, but there will be no data backing them so we're pretty safe.
	// We can create a cleanup operation that cleans up those orphaned records.
	if err := tx.Commit(); err != nil {
		return device, err
	}

	return device, nil
}

func (service *DeviceService) ActivateDevice(code, endpoint string) (models.Device, error) {
	var device models.Device

	var userId string
	if err := service.db.QueryRowxContext(service.ctx,
		`SELECT id
		 FROM users
		 WHERE idp_id = $1`,
		service.user,
	).Scan(&userId); err != nil {
		return device, err
	}

	var deviceId string
	var status string
	if err := service.db.QueryRowxContext(service.ctx,
		`SELECT id, status
		 FROM devices
		 WHERE code = $1
		 AND user_id = $2`,
		code,
		userId,
	).Scan(&deviceId, &status); err != nil {
		return device, err
	}

	if status == string(models.DeviceStatusActive) {
		return device, errors.New("device already activated")
	}

	result := service.authclient.Authorize(
		authz.PermissionActivateDevice,
		authz.Resource{Type: "device", ID: deviceId},
	)

	if result.Error != nil {
		return device, result.Error
	}

	if !result.Authorized {
		return device, authz.NewAuthzError(
			authz.PermissionActivateDevice,
			authz.Resource{Type: "device", ID: deviceId},
		)
	}

	if err := service.db.QueryRowxContext(
		service.ctx,
		`UPDATE devices
		 SET
		 	status = $1,
			endpoint = $2,
			modified = $3,
			modified_by = $4
		 WHERE id = $5
		 AND code = $6
		 AND (status = 'PENDING' OR status = 'INACTIVE')
		 RETURNING id, name, endpoint, user_id, status, code, created, created_by, modified, modified_by;`,
		models.DeviceStatusActive,
		endpoint,
		time.Now(),
		service.user,
		deviceId,
		code,
	).StructScan(&device); err != nil {
		return device, err
	}

	return device, nil
}

func (service *DeviceService) DeactivateDevice(id string) (models.Device, error) {
	var device models.Device

	var status string
	if err := service.db.QueryRowxContext(service.ctx,
		`SELECT status
		 FROM devices
		 WHERE id = $1`,
		id,
	).Scan(&status); err != nil {
		return device, err
	}

	if status == string(models.DeviceStatusInactive) || status == string(models.DeviceStatusPending) {
		return device, errors.New("device already inactive or pending")
	}

	result := service.authclient.Authorize(
		authz.PermissionDeactivateDevice,
		authz.Resource{Type: "device", ID: id},
	)

	if result.Error != nil {
		return device, result.Error
	}

	if !result.Authorized {
		return device, authz.NewAuthzError(
			authz.PermissionDeactivateDevice,
			authz.Resource{Type: "device", ID: id},
		)
	}

	if err := service.db.QueryRowxContext(
		service.ctx,
		`UPDATE devices
		 SET
		 	status = $1,
			modified = $2,
			modified_by = $3
		 WHERE id = $4
		 AND status = 'ACTIVE'
		 RETURNING id, name, endpoint, user_id, status, code, created, created_by, modified, modified_by;`,
		models.DeviceStatusInactive,
		time.Now(),
		service.user,
		id,
	).StructScan(&device); err != nil {
		return device, err
	}

	return device, nil
}

func (service *DeviceService) ListDevices(pagination DevicePagination) ([]models.Device, error) {
	results, err := service.listDevicesDataloader.Load(service.ctx, pagination)()
	return results, err
}

// deviceCodeCharacterSet represents the set of characters for generating a
// random device code.
var deviceCodeCharacterSet = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

// generateDeviceCode generates a random device code of n length using a
// predefined letterset.
func generateDeviceCode(n int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]rune, n)
	for i := range b {
		b[i] = deviceCodeCharacterSet[r.Intn(len(deviceCodeCharacterSet))]
	}
	return string(b)
}
