package main

type userStatus string

const (
	userStatusPending  userStatus = "PENDING"
	userStatusActive   userStatus = "ACTIVE"
	userStatusInactive userStatus = "INACTIVE"
)

// user represents a user user in the system.
type user struct {
	auditable
	ID    string `json:"id" db:"id"`
	Email string `json:"email" db:"email"`
	// IdpID is the ID of the user from their identity provider. Typically this
	// comes in the form of a sub claim of an identity token.
	IdpID  string     `json:"idpId" db:"idp_id"`
	Status userStatus `json:"status" db:"status"`
}
