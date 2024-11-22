package models

// A Page is the representation of an incident belonging to one or more
// agencies.
//
// Events are emitted during the lifetime of the Page, triggering notifications
// on member devices. A Page contains details about the incident and serves as
// a collaboration point for responding members.
type Page struct {
	Auditable
	Title       string   `dynamodbav:"name"`
	Description string   `dynamodbav:"description"`
	Location    Location `dynamodbav:"location"`
}
