package identity

// User represents a user. This representation is used for authorization
// decisions.
type User struct {
	Email       string                `json:"email"`
	IPDID       string                `json:"idpId"`
	Memberships map[string]Membership `json:"memberships"`
}

// Membership represents a binding between a user and an Agency.
type Membership struct {
	AgencyID string `json:"agencyId"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}
