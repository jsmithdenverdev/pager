package app

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/jsmithdenverdev/pager/pkg/identity"
)

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

// membership represents a users membership in an agency including their role.
type membership struct {
	PK         string           `dynamodbav:"pk"`
	SK         string           `dynamodbav:"sk"`
	Type       string           `dynamodbav:"type"`
	Status     membershipStatus `dynamodbav:"status"`
	Role       identity.Role    `dynamodbav:"role"`
	Created    time.Time        `dynamodbav:"created"`
	Modified   time.Time        `dynamodbav:"modified"`
	CreatedBy  string           `dynamodbav:"createdBy"`
	ModifiedBy string           `dynamodbav:"modifiedBy"`
}

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

type createMembershipRequest struct {
	AgencyID string        `json:"agencyId"`
	UserID   string        `json:"userId"`
	Role     identity.Role `json:"role"`
}

func (r createMembershipRequest) valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	if r.AgencyID == "" {
		problems["AgencyID"] = "agencyId is required"
	}

	if r.UserID == "" {
		problems["UserID"] = "userId is required"
	}

	if r.Role == "" {
		problems["Role"] = "role is required"
	}

	if !slices.Contains([]string{identity.RoleReader, identity.RoleWriter}, r.Role) {
		problems["Role"] = "role must be reader or writer"
	}

	return problems
}

type createMembershipResponse struct {
	AgencyID   string           `json:"agencyId"`
	UserID     string           `json:"userId"`
	Role       identity.Role    `json:"role"`
	Status     membershipStatus `json:"status"`
	Created    time.Time        `json:"created"`
	Modified   time.Time        `json:"modified"`
	CreatedBy  string           `json:"createdBy"`
	ModifiedBy string           `json:"modifiedBy"`
}

// listResponse represents a list of items with pagination.
type listResponse[T any] struct {
	Results     []T    `json:"results"`
	NextCursor  string `json:"nextCursor"`
	HasNextPage bool   `json:"hasNextPage"`
}
