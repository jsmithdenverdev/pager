package authz

type Resource struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}
