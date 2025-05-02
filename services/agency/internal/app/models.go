package app

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/jsmithdenverdev/pager/pkg/identity"
)

//-----------------------------------------------------------------------------
// AGENCY
//-----------------------------------------------------------------------------

// agency represents an agency in the database.
type agency struct {
	PK         string       `dynamodbav:"pk"`
	SK         string       `dynamodbav:"sk"`
	Type       entityType   `dynamodbav:"type"`
	Name       string       `dynamodbav:"name"`
	Status     agencyStatus `dynamodbav:"status"`
	Created    time.Time    `dynamodbav:"created"`
	Modified   time.Time    `dynamodbav:"modified"`
	CreatedBy  string       `dynamodbav:"createdBy"`
	ModifiedBy string       `dynamodbav:"modifiedBy"`
}

// agencyResponse represents a single agency by ID.
type agencyResponse struct {
	ID         string       `json:"pk"`
	Name       string       `json:"name"`
	Status     agencyStatus `json:"status"`
	Created    time.Time    `json:"created"`
	Modified   time.Time    `json:"modified"`
	CreatedBy  string       `json:"createdBy"`
	ModifiedBy string       `json:"modifiedBy"`
}

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

// toAgencyResponse converts an agency to a response.
func toAgencyResponse(agency agency) agencyResponse {
	return agencyResponse{
		ID:         strings.Split(agency.PK, "#")[1],
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

// membership represents a users membership in an agency including their role.
type membership struct {
	PK         string           `dynamodbav:"pk"`
	SK         string           `dynamodbav:"sk"`
	Type       entityType       `dynamodbav:"type"`
	Status     membershipStatus `dynamodbav:"status"`
	Role       identity.Role    `dynamodbav:"role"`
	Created    time.Time        `dynamodbav:"created"`
	Modified   time.Time        `dynamodbav:"modified"`
	CreatedBy  string           `dynamodbav:"createdBy"`
	ModifiedBy string           `dynamodbav:"modifiedBy"`
}

// membershipResponse represents a single membership by ID.
type membershipResponse struct {
	AgencyID   string           `json:"agencyId"`
	UserID     string           `json:"userId"`
	Role       identity.Role    `json:"role"`
	Status     membershipStatus `json:"status"`
	Created    time.Time        `json:"created"`
	Modified   time.Time        `json:"modified"`
	CreatedBy  string           `json:"createdBy"`
	ModifiedBy string           `json:"modifiedBy"`
}

//-----------------------------------------------------------------------------
// INVITATION
//-----------------------------------------------------------------------------

// invitation represents an invitation to join an agency.
type invitation struct {
	PK         string           `dynamodbav:"pk"`
	SK         string           `dynamodbav:"sk"`
	Type       entityType       `dynamodbav:"type"`
	Status     invitationStatus `dynamodbav:"status"`
	Role       identity.Role    `dynamodbav:"role"`
	Created    time.Time        `dynamodbav:"created"`
	Modified   time.Time        `dynamodbav:"modified"`
	CreatedBy  string           `dynamodbav:"createdBy"`
	ModifiedBy string           `dynamodbav:"modifiedBy"`
}

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
	AgencyID   string           `json:"agencyId"`
	Email      string           `json:"email"`
	Role       identity.Role    `json:"role"`
	Status     invitationStatus `json:"status"`
	Created    time.Time        `json:"created"`
	Modified   time.Time        `json:"modified"`
	CreatedBy  string           `json:"createdBy"`
	ModifiedBy string           `json:"modifiedBy"`
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
