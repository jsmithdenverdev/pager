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
// FIELDS
//-----------------------------------------------------------------------------

type keyFields struct {
	PK   string     `dynamodbav:"pk"`
	SK   string     `dynamodbav:"sk"`
	Type entityType `dynamodbav:"type"`
}

type auditableFields struct {
	Created    time.Time `dynamodbav:"created"`
	Modified   time.Time `dynamodbav:"modified"`
	CreatedBy  string    `dynamodbav:"createdBy"`
	ModifiedBy string    `dynamodbav:"modifiedBy"`
}

func newAuditableFields(userID string, timestamp time.Time) auditableFields {
	return auditableFields{
		Created:    timestamp,
		Modified:   timestamp,
		CreatedBy:  userID,
		ModifiedBy: userID,
	}
}

//-----------------------------------------------------------------------------
// ENDPOINT
//-----------------------------------------------------------------------------

// endpoint represents an endpoint that can be used to send notifications.
// Endpoints are registered to an agency.
type endpoint struct {
	keyFields
	auditableFields
	EndpointType  endpointType `dynamodbav:"endpointType"`
	Name          string       `dynamodbav:"name"`
	URL           string       `dynamodbav:"url"`
	Registrations []string     `dynamodbav:"registrations"`
	UserID        string       `dynamodbav:"userId"`
}

type endpointResponse struct {
	ID            string       `json:"id"`
	UserID        string       `json:"userId"`
	EndpointType  endpointType `json:"endpointType"`
	Name          string       `json:"name"`
	URL           string       `json:"url"`
	Registrations []string     `json:"registrations"`
	Created       time.Time    `json:"created"`
	Modified      time.Time    `json:"modified"`
	CreatedBy     string       `json:"createdBy"`
	ModifiedBy    string       `json:"modifiedBy"`
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
	keyFields
	auditableFields
	EndpointID string `dynamodbav:"endpointId"`
	UserID     string `dynamodbav:"userId"`
}

//-----------------------------------------------------------------------------
// OWNERSHIP LINK
//-----------------------------------------------------------------------------

// ownershipLink represents the ownershipLink of an endpoint by a user.
// The model is a simple relationship binding that doesn't include other
// metadata. The relationship is encoded within the pk and sk.
type ownershipLink struct {
	keyFields
	auditableFields
}

type ownershipLinkResponse struct {
	UserID     string `json:"userId"`
	EndpointID string `json:"endpointId"`
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
