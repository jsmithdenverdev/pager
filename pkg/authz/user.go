package authz

// UserAgency represents information about a user's agency needed to make
// authorization decisions.
type UserAgency struct {
	ID      string `json:"id"`
	Role    string `json:"role"`
	Devices []string
}

// UserDevice represents information about a user's device needed to make
// authorization decisions.
type UserDevice struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// User represents information about a user needed to make authorization
// decisions.
type User struct {
	Email        string                `json:"email"`
	Entitlements []Entitlement         `json:"entitlements"`
	IPDID        string                `json:"idpId"`
	Status       string                `json:"status"`
	ActiveAgency string                `json:"activeAgency"`
	Agencies     map[string]UserAgency `json:"agencies"`
	Devices      map[string]UserDevice `json:"devices"`
}
