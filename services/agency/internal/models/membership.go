package models

import "time"

// Membership represents a relationship between an agency and a user. This
// allows a user to query the agencies they are a member of.
type Membership struct {
	Model
	IDPID          string    `dynamodbav:"idpid"`
	AgencyName     string    `dynamodbav:"name"`
	AgencyCreated  time.Time `dynamodbav:"agency_created"`
	AgencyModified time.Time `dynamodbav:"agency_modified"`
}
