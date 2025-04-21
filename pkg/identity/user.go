package identity

// User represents a user. This representation is used for authorization
// decisions.
type User struct {
	Email string `json:"email"`
	IPDID string `json:"idpId"`
}
