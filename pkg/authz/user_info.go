package authz

// UserInfo represents information about a user needed to make authorization
// decisions.
type UserInfo struct {
	Email        string        `json:"emai"`
	Entitlements []Entitlement `json:"entitlements"`
	IPDID        string        `json:"ipdId"`
	Status       string        `json:"status"`
	Accounts     map[string]struct {
		Roles []string `json:"roles"`
	} `json:"accounts"`
}
