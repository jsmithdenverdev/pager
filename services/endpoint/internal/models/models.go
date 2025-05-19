package models

import (
	"time"
)

type EndpointType = string

const (
	EndpointTypePush EndpointType = "PUSH"
)

type KeyFields struct {
	PK   string     `dynamodbav:"pk"`
	SK   string     `dynamodbav:"sk"`
	Type EntityType `dynamodbav:"type"`
}

type AuditableFields struct {
	Created    time.Time `dynamodbav:"created"`
	Modified   time.Time `dynamodbav:"modified"`
	CreatedBy  string    `dynamodbav:"createdBy"`
	ModifiedBy string    `dynamodbav:"modifiedBy"`
}

func NewAuditableFields(userID string, timestamp time.Time) AuditableFields {
	return AuditableFields{
		Created:    timestamp,
		Modified:   timestamp,
		CreatedBy:  userID,
		ModifiedBy: userID,
	}
}
