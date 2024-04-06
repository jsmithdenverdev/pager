package schema

import goplaygroundvalidator "github.com/go-playground/validator/v10"

// validator is a global singleton validation service. It is stateless and safe
// to access from tests.
var validator = goplaygroundvalidator.New(
	goplaygroundvalidator.WithRequiredStructEnabled(),
)
