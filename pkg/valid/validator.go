package valid

import "context"

// Validator defines a method for validating an object. It returns a slice of
// problems found during validation.
type Validator interface {
	Valid(ctx context.Context) (problems []Problem)
}
