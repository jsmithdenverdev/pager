package models

// DeviceRegistration represents a relationship between an agency and a device.
// This allows us to query the devices belonging to an agency.
type DeviceRegistration struct {
	Model
	DeviceID string `dynamodbav:"device_id"`
	Active   bool   `dynamodbav:"active"`
}
