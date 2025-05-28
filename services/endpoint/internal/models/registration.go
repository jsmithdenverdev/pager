package models

type Registration struct {
	Type EntityType `dynamodbav:"type"`
	KeyFields
	AuditableFields
}
