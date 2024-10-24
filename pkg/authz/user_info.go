package authz

// UserInfo represents information about a user needed to make authorization
// decisions.
type UserInfo struct {
	Email        string        `json:"email"`
	Entitlements []Entitlement `json:"entitlements"`
	IPDID        string        `json:"idpId"`
	Status       string        `json:"status"`
	Accounts     map[string]struct {
		Role string `json:"role"`
	} `json:"accounts"`
	// If requests are being done within the context of an Agency, ActiveAgency
	// will be populated with that agency identifier.
	ActiveAgency string
}
