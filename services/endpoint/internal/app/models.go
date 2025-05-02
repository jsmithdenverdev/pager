package app

type endpointType = string

const (
	endpointTypePush endpointType = "push"
)

// endpoint represents an endpoint that can be used to send notifications.
// Endpoints are registered to an agency.
type endpoint struct {
	PK           string       `dynamodbav:"pk"`
	SK           string       `dynamodbav:"sk"`
	Type         string       `dynamodbav:"type"`
	UserID       string       `dynamodbav:"userId"`
	EndpointType endpointType `dynamodbav:"endpointType"`
	Name         string       `dynamodbav:"name"`
	URL          string       `dynamodbav:"url"`
}

type endpointResponse struct {
	ID           string       `json:"id"`
	AgencyID     string       `json:"agencyId"`
	UserID       string       `json:"userId,omitempty"`
	EndpointType endpointType `json:"endpointType,omitempty"`
	Name         string       `json:"name,omitempty"`
	URL          string       `json:"url,omitempty"`
}

// registration represents a registration of an endpoint to an account. The
// endpoint must be registered to an account before it can be used.
type registration struct {
	PK     string `dynamodbav:"pk"`
	SK     string `dynamodbav:"sk"`
	Type   string `dynamodbav:"type"`
	UserID string `dynamodbav:"userId"`
}

type registrationResponse struct {
	EndpointID string `json:"endpointId"`
	AccountID  string `json:"accountId"`
	UserID     string `json:"userId"`
}

// listResponse represents a list of items with pagination.
type listResponse[T any] struct {
	Results     []T    `json:"results"`
	NextCursor  string `json:"nextCursor"`
	HasNextPage bool   `json:"hasNextPage"`
}
