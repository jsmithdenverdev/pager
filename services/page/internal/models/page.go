package models

import "time"

// Page represents a Page in the database.
type Page struct {
	PK         string     `dynamodbav:"pk"`
	SK         string     `dynamodbav:"sk"`
	Type       EntityType `dynamodbav:"type"`
	Title      string     `dynamodbav:"title"`
	Notes      string     `dynamodbav:"notes"`
	Notify     bool       `dynamodbav:"notify"`
	Location   Location   `dynamodbav:"location"`
	Created    time.Time  `dynamodbav:"created"`
	Modified   time.Time  `dynamodbav:"modified"`
	CreatedBy  string     `dynamodbav:"createdBy"`
	ModifiedBy string     `dynamodbav:"modifiedBy"`
}

// Location represents a location for a page. The location may be a common name (e.g., "Kelso Ridge") or may be a
// set of coordinates following a specific type (e.g., decimal degrees, degrees minutes seconds, etc.).
type Location struct {
	Description string  `dynamodbav:"description"`
	Latitude    float64 `dynamodbav:"latitude"`
	Longitude   float64 `dynamodbav:"longitude"`
	Type        string  `dynamodbav:"type"`
}
