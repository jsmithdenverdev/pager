package app

import "context"

type createPageRequest struct {
	Agencies []string `json:"agencies"`
	Title    string   `json:"title"`
	Notes    string   `json:"notes"`
	Notify   bool     `json:"notify"`
	Location struct {
		Description string  `json:"description"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
		Type        string  `json:"type"`
	} `json:"location"`
}

func (r createPageRequest) valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	if len(r.Agencies) == 0 {
		problems["agencies"] = "must create page with at least one agency"
	}

	if r.Title == "" {
		problems["title"] = "page must have a title"
	}

	return problems
}

type createPageResponse struct {
	ID string `json:"id"`
}

// listResponse represents a list of items with pagination.
type listResponse[T any] struct {
	Results     []T    `json:"results"`
	NextCursor  string `json:"nextCursor"`
	HasNextPage bool   `json:"hasNextPage"`
}
