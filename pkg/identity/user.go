package identity

import "time"

// User represents a user. This representation is used for authorization
// decisions.
type User struct {
	ID           string          `json:"id"`
	Email        string          `json:"email"`
	Name         string          `json:"name"`
	Status       Status          `json:"status"`
	Entitlements []Entitlement   `json:"entitlements,omitempty"`
	Memberships  map[string]Role `json:"memberships,omitempty"`
	Created      time.Time       `json:"created"`
	Modified     time.Time       `json:"modified"`
	CreatedBy    string          `json:"createdBy"`
	ModifiedBy   string          `json:"modifiedBy"`
}
