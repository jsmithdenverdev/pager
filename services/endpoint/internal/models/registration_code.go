package models

// RegistrationCode represents a registration of an endpoint to an account. The
// endpoint must be registered to an account before it can be used.
type RegistrationCode struct {
	KeyFields
	AuditableFields
	EndpointID string `dynamodbav:"endpointId"`
	UserID     string `dynamodbav:"userId"`
}
