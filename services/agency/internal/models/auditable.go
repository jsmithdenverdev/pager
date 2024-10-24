package models

import "time"

type Auditable struct {
	PK         string    `dynamodbav:"pk"`
	SK         string    `dynamodbav:"sk"`
	ID         string    `dynamodbav:"id"`
	Created    time.Time `dynamodbav:"created"`
	CreatedBy  string    `dynamodbav:"created_by"`
	Modified   time.Time `dynamodbav:"modified"`
	ModifiedBy string    `dynamodbav:"modified_by"`
}

func (auditable Auditable) Identity() string {
	return auditable.ID
}
