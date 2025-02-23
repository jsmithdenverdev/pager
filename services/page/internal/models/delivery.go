package models

type DeliveryStatus string

const (
	DeliveryStatusPending  DeliveryStatus = "QUEUED"
	DeliveryStatusActive   DeliveryStatus = "DELIVERING"
	DeliveryStatusInactive DeliveryStatus = "DELIVERED"
	DeliveryStatusFailed   DeliveryStatus = "FAILED"
)

// A Page is the representation of an incident belonging to one or more
// agencies.
//
// Events are emitted during the lifetime of the Page, triggering notifications
// on member devices. A Page contains details about the incident and serves as
// a collaboration point for responding members.
type Delivery struct {
	Auditable
	DeviceID string         `dynamodbav:"device_id"`
	Status   DeliveryStatus `dynamodbav:"delivery_status"`
}
