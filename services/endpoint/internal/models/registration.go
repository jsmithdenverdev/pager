package models

type Registration struct {
	KeyFields
	AuditableFields
	Type         EntityType   `dynamodbav:"type"`
	URL          string       `dynamodbav:"url"`
	EndpointType EndpointType `dynamodbav:"endpointType"`
}
