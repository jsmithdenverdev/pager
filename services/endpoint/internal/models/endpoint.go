package models

// Endpoint represents an Endpoint that can be used to send notifications.
// Endpoints are registered to an agency.
type Endpoint struct {
	KeyFields
	AuditableFields
	EndpointType     EndpointType   `dynamodbav:"endpointType"`
	Name             string         `dynamodbav:"name"`
	URL              string         `dynamodbav:"url"`
	Registrations    map[string]any `dynamodbav:"registrations"`
	UserID           string         `dynamodbav:"userId"`
	RegistrationCode string         `dynamodbav:"registrationCode"`
}
