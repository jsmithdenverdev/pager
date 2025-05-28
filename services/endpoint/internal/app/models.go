package app

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/jsmithdenverdev/pager/services/endpoint/internal/models"
)

//-----------------------------------------------------------------------------
// ENDPOINT
//-----------------------------------------------------------------------------

type endpointResponse struct {
	ID               string         `json:"id"`
	UserID           string         `json:"userId"`
	EndpointType     string         `json:"endpointType"`
	Name             string         `json:"name"`
	URL              string         `json:"url"`
	Registrations    map[string]any `json:"registrations"`
	RegistrationCode string         `json:"registrationCode"`
	Created          time.Time      `json:"created"`
	Modified         time.Time      `json:"modified"`
	CreatedBy        string         `json:"createdBy"`
	ModifiedBy       string         `json:"modifiedBy"`
}

func toEndpointResponse(endpoint models.Endpoint) endpointResponse {
	return endpointResponse{
		ID:               strings.Split(endpoint.PK, "#")[1],
		UserID:           endpoint.UserID,
		EndpointType:     endpoint.EndpointType,
		Name:             endpoint.Name,
		URL:              endpoint.URL,
		Registrations:    endpoint.Registrations,
		RegistrationCode: endpoint.RegistrationCode,
		Created:          endpoint.Created,
		Modified:         endpoint.Modified,
		CreatedBy:        endpoint.CreatedBy,
		ModifiedBy:       endpoint.ModifiedBy,
	}
}

type createEndpointRequest struct {
	URL          string `json:"url"`
	Name         string `json:"name"`
	EndpointType string `json:"endpointType"`
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

	if !slices.Contains([]models.EndpointType{models.EndpointTypePush}, r.EndpointType) {
		problems["endpointType"] = fmt.Sprintf("endpointType must be one of: %s", strings.Join([]models.EndpointType{models.EndpointTypePush}, ", "))
	}

	return problems
}

type createEndpointResponse struct {
	ID string `json:"id"`
}

//-----------------------------------------------------------------------------
// OWNER
//-----------------------------------------------------------------------------

type ownerResponse struct {
	UserID     string `json:"userId"`
	EndpointID string `json:"endpointId"`
}

func toOwnerResponse(link models.Owner) ownerResponse {
	return ownerResponse{
		UserID:     strings.Split(link.PK, "#")[1],
		EndpointID: strings.Split(link.SK, "#")[1],
	}
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
