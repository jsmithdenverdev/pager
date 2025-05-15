package models

import (
	"time"
)

type EntityType string

const (
	EntityTypeUser       EntityType = "USER"
	EntityTypeUserLookup EntityType = "USER_LOOKUP"
)

type Keys struct {
	PK   string     `dynamodbav:"pk"`
	SK   string     `dynamodbav:"sk"`
	Type EntityType `dynamodbav:"type"`
}

type Auditable struct {
	Created    time.Time `dynamodbav:"created"`
	Modified   time.Time `dynamodbav:"modified"`
	CreatedBy  string    `dynamodbav:"createdBy"`
	ModifiedBy string    `dynamodbav:"modifiedBy"`
}

func NewAuditable(userID string, timestamp time.Time) Auditable {
	return Auditable{
		Created:    timestamp,
		Modified:   timestamp,
		CreatedBy:  userID,
		ModifiedBy: userID,
	}
}
