package service

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"math/rand"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/jsmithdenverdev/pager/authz"
	"github.com/jsmithdenverdev/pager/models"
)

// DeviceService exposes all operations that can be performed on or for devices.
type DeviceService struct {
	ctx        context.Context
	user       string
	authclient *authz.Client
	db         *sqlx.DB
	logger     *slog.Logger
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
		ctx:        ctx,
		user:       user,
		authclient: authz,
		db:         db,
		logger:     logger,
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
		`SELECT idp_id
		 FROM users
		 WHERE id = $1`,
		ownerId,
	).Scan(&ownerIdpID); err != nil {
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
