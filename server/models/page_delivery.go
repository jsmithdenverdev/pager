package models

type PageDeliveryStatus string

const (
	PageDeliveryStatusPending        PageDeliveryStatus = "PENDING"
	PageDeliveryStatusDelivering     PageDeliveryStatus = "DELIVERING"
	PageDeliveryStatusDelivered      PageDeliveryStatus = "DELIVERED"
	PageDeliveryStatusDeliveryFailed PageDeliveryStatus = "DELIVERY_FAILED"
)

// PageDelivery is a core entity of pager.
//
// A page delivery represents binding between a page and a device, as well as a
// status denoting if the page has actually been delivered.
type PageDelivery struct {
	Auditable
	PageID   string             `json:"pageId" db:"page_id"`
	DeviceID string             `json:"deviceId" db:"device_id"`
	Status   PageDeliveryStatus `json:"status" db:"status"`
}
