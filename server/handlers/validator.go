package handlers

import "context"

// validator defines a method for validating an object. It returns a slice of
// problems found during validation.
type validator interface {
	Valid(ctx context.Context) (problems []problem)
}

// problem represents an issue found during validation.
type problem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
