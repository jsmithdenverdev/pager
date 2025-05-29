package models

import "time"

type AgencyStatus = string

const (
	AgencyStatusActive   AgencyStatus = "ACTIVE"
	AgencyStatusInactive AgencyStatus = "INACTIVE"
)

// agency represents an agency in the database.
type Agency struct {
	PK         string       `dynamodbav:"pk"`
	SK         string       `dynamodbav:"sk"`
	Type       EntityType   `dynamodbav:"type"`
	Name       string       `dynamodbav:"name"`
	Status     AgencyStatus `dynamodbav:"status"`
	Created    time.Time    `dynamodbav:"created"`
	Modified   time.Time    `dynamodbav:"modified"`
	CreatedBy  string       `dynamodbav:"createdBy"`
	ModifiedBy string       `dynamodbav:"modifiedBy"`
}
