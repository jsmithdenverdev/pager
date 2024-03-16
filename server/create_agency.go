package main

import (
	"fmt"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/graphql-go/graphql"
)

// createAgencyInput represents the fields needed to create a new agency.
type createAgencyInput struct {
	Name string `json:"name"`
}

// createAgencyInputType is the graphql input type for the createAgency
// mutation.
var createAgencyInputType = graphql.InputObjectConfig{
	Name: "CreateAgencyInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"name": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
	},
}

// createAgencyResultType is the graphql result type for the createAgency
// mutation.
var createAgencyResultType = newResultType[agency](
	"CreateAgencyResult",
	agencyType(),
	baseErrorType,
	validationErrorType,
	authzErrorType)

// toCreateAgencyInput converts a `map[string]interface{}` into a
// `createAgencyInput`.
func toCreateAgencyInput(args map[string]interface{}) createAgencyInput {
	var input createAgencyInput
	// Name
	name, ok := args["name"].(string)
	if !ok {
		name = ""
	}
	input.Name = name
	return input
}

// createAgencyMutation returns the `createAgency` mutation field.
//
// `createAgency` allows a pager admin to create a new agency in the system on
// behalf of a real world agency. The agency is created in an `INACTIVE` status
func createAgencyMutation(logger *slog.Logger, validate *validator.Validate) *graphql.Field {
	return &graphql.Field{
		Name: "createAgency",
		Type: createAgencyResultType,
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewInputObject(createAgencyInputType),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := toCreateAgencyInput(p.Args["input"].(map[string]interface{}))
			if err := validate.Struct(input); err != nil {
				return err, nil
			}
			logger.InfoContext(p.Context, "createAgency", "input", fmt.Sprintf("%+v", p.Args["input"]))
			return newAuthzError("user", "system", "create-agency"), nil
		},
	}
}
