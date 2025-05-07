package app

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"
)

type endpointType = string

const (
	endpointTypePush endpointType = "push"
)

//-----------------------------------------------------------------------------
// ENDPOINT
//-----------------------------------------------------------------------------

// endpoint represents an endpoint that can be used to send notifications.
// Endpoints are registered to an agency.
type endpoint struct {
	PK            string       `dynamodbav:"pk"`
	SK            string       `dynamodbav:"sk"`
	Type          entityType   `dynamodbav:"type"`
	EndpointType  endpointType `dynamodbav:"endpointType"`
	Name          string       `dynamodbav:"name"`
	URL           string       `dynamodbav:"url"`
	Registrations []string     `dynamodbav:"registrations"`
	Created       time.Time    `dynamodbav:"created"`
	Modified      time.Time    `dynamodbav:"modified"`
	CreatedBy     string       `dynamodbav:"createdBy"`
	ModifiedBy    string       `dynamodbav:"modifiedBy"`
}

type endpointResponse struct {
	ID            string       `json:"id"`
	UserID        string       `json:"userId,omitempty"`
	EndpointType  endpointType `json:"endpointType,omitempty"`
	Name          string       `json:"name,omitempty"`
	URL           string       `json:"url,omitempty"`
	Registrations []string     `json:"registrations,omitempty"`
}

type createEndpointRequest struct {
	URL          string       `json:"url"`
	Name         string       `json:"name"`
	EndpointType endpointType `json:"endpointType"`
}

func (r createEndpointRequest) valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	if r.URL == "" {
		problems["url"] = "url is required"
	}

	if r.Name == "" {
		problems["name"] = "name is required"
	}

	if r.EndpointType == "" {
		problems["endpointType"] = "endpointType is required"
	}

	if !slices.Contains([]endpointType{endpointTypePush}, r.EndpointType) {
		problems["endpointType"] = fmt.Sprintf("endpointType must be one of %s", strings.Join([]endpointType{endpointTypePush}, ", "))
	}

	return problems
}

type createEndpointResponse struct {
	ID string `json:"id"`
}

//-----------------------------------------------------------------------------
// REGISTRATION CODE
//-----------------------------------------------------------------------------

// registrationCode represents a registration of an endpoint to an account. The
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

//-----------------------------------------------------------------------------
// LIST RESPONSE
//-----------------------------------------------------------------------------

// listResponse represents a list of items with pagination.
type listResponse[T any] struct {
	Results     []T    `json:"results"`
	NextCursor  string `json:"nextCursor"`
	HasNextPage bool   `json:"hasNextPage"`
}
