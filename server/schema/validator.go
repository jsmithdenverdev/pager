package schema

import (
	goplaygroundvalidator "github.com/go-playground/validator/v10"
)

var validator = goplaygroundvalidator.New(goplaygroundvalidator.WithRequiredStructEnabled())
