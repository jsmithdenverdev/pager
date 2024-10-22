package authz

// UserInfo represents information about a user needed to make authorization
// decisions.
type UserInfo struct {
	Email        string        `json:"email"`
	Entitlements []Entitlement `json:"entitlements"`
	IPDID        string        `json:"idpId"`
	Status       string        `json:"status"`
	Accounts     map[string]struct {
		Roles []string `json:"roles"`
	} `json:"accounts"`
}
