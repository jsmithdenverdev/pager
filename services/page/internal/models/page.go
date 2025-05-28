package models

import "time"

// Page represents a Page in the database.
type Page struct {
	PK         string     `dynamodbav:"pk"`
	SK         string     `dynamodbav:"sk"`
	Type       EntityType `dynamodbav:"type"`
	Title      string     `dynamodbav:"title"`
	Notes      string     `dynamodbav:"notes"`
	Created    time.Time  `dynamodbav:"created"`
	Modified   time.Time  `dynamodbav:"modified"`
	CreatedBy  string     `dynamodbav:"createdBy"`
	ModifiedBy string     `dynamodbav:"modifiedBy"`
}
