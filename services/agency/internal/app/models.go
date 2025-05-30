package app

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/jsmithdenverdev/pager/pkg/identity"
	"github.com/jsmithdenverdev/pager/services/agency/internal/models"
)

//-----------------------------------------------------------------------------
// AGENCY
//-----------------------------------------------------------------------------

// createAgencyRequest represents a request to create a new agency.
type createAgencyRequest struct {
	Name string `json:"name"`
}

// valid returns a map of validation problems for the request.
func (r createAgencyRequest) valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	if r.Name == "" {
		problems["Name"] = "name is required"
	}

	return problems
}

// createAgencyResponse represents a response to a request to create a new
// agency.
type createAgencyResponse struct {
	ID string `json:"id"`
}

// agencyResponse represents a single agency by ID.
type agencyResponse struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	Created    time.Time `json:"created"`
	Modified   time.Time `json:"modified"`
	CreatedBy  string    `json:"createdBy"`
	ModifiedBy string    `json:"modifiedBy"`
}

// toAgencyResponse converts an agency to a response.
func toAgencyResponse(agency models.Agency) agencyResponse {
	return agencyResponse{
		ID:         agency.ID,
		Name:       agency.Name,
		Status:     agency.Status,
		Created:    agency.Created,
		Modified:   agency.Modified,
		CreatedBy:  agency.CreatedBy,
		ModifiedBy: agency.ModifiedBy,
	}
}

//-----------------------------------------------------------------------------
// MEMBERSHIP
//-----------------------------------------------------------------------------

// membershipResponse represents a single membership by ID.
type membershipResponse struct {
	AgencyID   string        `json:"agencyId"`
	UserID     string        `json:"userId"`
	Role       identity.Role `json:"role"`
	Status     string        `json:"status"`
	Created    time.Time     `json:"created"`
	Modified   time.Time     `json:"modified"`
	CreatedBy  string        `json:"createdBy"`
	ModifiedBy string        `json:"modifiedBy"`
}

//-----------------------------------------------------------------------------
// INVITATION
//-----------------------------------------------------------------------------

// createInvitationRequest represents a request to create a new invitation.
type createInvitationRequest struct {
	Email string        `json:"email"`
	Role  identity.Role `json:"role"`
}

// valid returns a map of validation problems for the request.
func (r createInvitationRequest) valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	if r.Email == "" {
		problems["email"] = "email is required"
	}

	if r.Role == "" {
		problems["role"] = "role is required"
	}

	validRoles := []identity.Role{identity.RoleReader, identity.RoleWriter}

	if !slices.Contains(validRoles, r.Role) {
		problems["role"] = fmt.Sprintf("role must be one of %s", strings.Join(validRoles, ", "))
	}

	return problems
}

// createInvitationResponse represents a response to a request to create a new
// invitation.
type createInvitationResponse struct {
	AgencyID   string        `json:"agencyId"`
	Email      string        `json:"email"`
	Role       identity.Role `json:"role"`
	Status     string        `json:"status"`
	Created    time.Time     `json:"created"`
	Modified   time.Time     `json:"modified"`
	CreatedBy  string        `json:"createdBy"`
	ModifiedBy string        `json:"modifiedBy"`
}

//-----------------------------------------------------------------------------
// ENDPOINT REGISTRATION
//-----------------------------------------------------------------------------

// registerEndpointRequest represents a request to create a new registration.
type registerEndpointRequest struct {
	RegistrationCode string `json:"registrationCode"`
}

// valid returns a map of validation problems for the request.
func (r registerEndpointRequest) valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	if r.RegistrationCode == "" {
		problems["registrationCode"] = "registrationCode is required"
	}

	return problems
}

// registerEndpointResponse represents a response to a request to create a new
// registration.
type registerEndpointResponse struct {
	AgencyID         string    `json:"agencyId"`
	RegistrationCode string    `json:"registrationCode"`
	Status           string    `json:"status"`
	Created          time.Time `json:"created"`
	Modified         time.Time `json:"modified"`
	CreatedBy        string    `json:"createdBy"`
	ModifiedBy       string    `json:"modifiedBy"`
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
