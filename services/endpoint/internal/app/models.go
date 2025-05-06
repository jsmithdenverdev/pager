package app

import "strings"

type endpointType = string

const (
	endpointTypePush endpointType = "push"
)

// endpoint represents an endpoint that can be used to send notifications.
// Endpoints are registered to an agency.
type endpoint struct {
	PK            string       `dynamodbav:"pk"`
	SK            string       `dynamodbav:"sk"`
	Type          string       `dynamodbav:"type"`
	EndpointType  endpointType `dynamodbav:"endpointType"`
	Name          string       `dynamodbav:"name"`
	URL           string       `dynamodbav:"url"`
	Registrations []string     `dynamodbav:"registrations"`
}

type endpointResponse struct {
	ID            string       `json:"id"`
	UserID        string       `json:"userId,omitempty"`
	EndpointType  endpointType `json:"endpointType,omitempty"`
	Name          string       `json:"name,omitempty"`
	URL           string       `json:"url,omitempty"`
	Registrations []string     `json:"registrations,omitempty"`
}

// registration represents a registration of an endpoint to an account. The
// endpoint must be registered to an account before it can be used.
type registrationCode struct {
	PK string `dynamodbav:"pk"`
	SK string `dynamodbav:"sk"`
}

// EndpointID returns the endpoint ID from the registration code.
func (r registrationCode) EndpointID() string {
	return strings.Split(r.SK, "#")[1]
}

// UserID returns the user ID from the registration code.
func (r registrationCode) UserID() string {
	return strings.Split(r.PK, "#")[3]
}

// listResponse represents a list of items with pagination.
type listResponse[T any] struct {
	Results     []T    `json:"results"`
	NextCursor  string `json:"nextCursor"`
	HasNextPage bool   `json:"hasNextPage"`
}
