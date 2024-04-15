package models

type DeviceStatus string

const (
	DeviceStatusPending  DeviceStatus = "PENDING"
	DeviceStatusActive   DeviceStatus = "ACTIVE"
	DeviceStatusInactive DeviceStatus = "INACTIVE"
)

// Device is a core entity of pager.
//
// A device belongs to a User, who may be registered with one or more Agencies.
// Notifications in pager are pushed to Devices. Currently, push notifications
// are what will be delivered, and we track the push endpoint for that device
// on the row.
type Device struct {
	Auditable
	Name     string       `json:"name" db:"name"`
	Status   DeviceStatus `json:"status" db:"status"`
	Endpoint string       `json:"endpoint" db:"endpoint"`
	UserID   string       `json:"userId" db:"user_id"`
	Code     string       `json:"code" db:"code"`
}
