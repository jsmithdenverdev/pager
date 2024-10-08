package models

type UserStatus string

const (
	UserStatusPending  UserStatus = "PENDING"
	UserStatusActive   UserStatus = "ACTIVE"
	UserStatusInactive UserStatus = "INACTIVE"
)

// User represents a User in the system.
//
// A User belongs to one or more Agencies.
type User struct {
	Auditable
	Email string `db:"email"`
	// IdpID is the ID of the user from their identity provider. Typically this
	// comes in the form of a sub claim of an identity token.
	IdpID  string     `db:"idp_id"`
	Status UserStatus `db:"status"`
}
