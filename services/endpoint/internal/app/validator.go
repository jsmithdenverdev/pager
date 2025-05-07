package app

import "context"

// validator is an object that can be validated.
type validator interface {
	// valid checks the object and returns any
	// problems. If len(problems) == 0 then
	// the object is valid.
	valid(ctx context.Context) (problems map[string]string)
}
