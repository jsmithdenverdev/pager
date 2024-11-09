package models

import "time"

type Auditable struct {
	Created    time.Time `dynamodbav:"created"`
	CreatedBy  string    `dynamodbav:"created_by"`
	Modified   time.Time `dynamodbav:"modified"`
	ModifiedBy string    `dynamodbav:"modified_by"`
}

type Model struct {
	Auditable
	PK   string `dynamodbav:"pk"`
	SK   string `dynamodbav:"sk"`
	Type string `dynamodbav:"type"`
}
