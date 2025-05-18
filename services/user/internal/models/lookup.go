package models

// Lookup is a record used to Lookup a user by various attributes. It consists
// of a pk which is the type of Lookup (e.g. email) and the users id.
type Lookup struct {
	Keys
	Auditable
	UserID string `dynamodbav:"userId"`
}
