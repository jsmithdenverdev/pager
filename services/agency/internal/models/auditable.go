package models

import "time"

type Auditable struct {
	PK         string    `dynamodbav:"pk"`
	SK         string    `dynamodbav:"sk"`
	Created    time.Time `dynamodbav:"created"`
	CreatedBy  string    `dynamodbav:"created_by"`
	Modified   time.Time `dynamodbav:"modified"`
	ModifiedBy string    `dynamodbav:"modified_by"`
}
