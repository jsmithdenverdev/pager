package handlers

import (
	"context"

	"github.com/jsmithdenverdev/pager/models"
)

// createAgencyRequest represents the data required to create a new Agency.
type createAgencyRequest struct {
	Name string `json:"name"`
}

// Valid performs validations on a createAgencyRequest and returns a slice of
// problem structs if issues are encountered.
func (r createAgencyRequest) Valid(ctx context.Context) []problem {
	var problems []problem
	if r.Name == "" {
		problems = append(problems, problem{
			Name:        "name",
			Description: "Name must be at least 1 character",
		})
	}
	return problems
}

// MapTo maps a createAgencyRequest to a models.Agency.
func (r createAgencyRequest) MapTo() models.Agency {
	var m models.Agency
	m.Name = r.Name
	return m
}
