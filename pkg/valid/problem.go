package valid

// Problem represents an issue found during validation.
type Problem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
