package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/jsmithdenverdev/pager/models"
	"github.com/jsmithdenverdev/pager/service"
)

// createAgencyInput represents the fields needed to create a new agency.
type createAgencyInput struct {
	Name string `json:"name" validate:"min=1"`
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

// toCreateAgencyInput converts a `map[string]interface{}` into a
// `createAgencyInput` performing validation on the model and returning any
// errors.
func toCreateAgencyInput(args map[string]interface{}) (createAgencyInput, error) {
	var input createAgencyInput
	// Name
	name, ok := args["name"].(string)
	if !ok {
		name = ""
	}
	input.Name = name
	return input, validator.Struct(input)
}

// createAgencyPayload is the struct representation of the result of a
// successful CreateAgency mutation.
type createAgencyPayload struct {
	Agency models.Agency `json:"agency"`
}

// createAgencyPayloadType is the graphql representation of the result of a
// successful CreateAgency mutation.
var createAgencyPayloadType = graphql.NewObject(graphql.ObjectConfig{
	Name: "CreateAgencyPayload",
	Fields: graphql.Fields{
		"agency": &graphql.Field{
			Type: agencyType,
		},
	},
})

// createAgencyMutation is the field definition for the createAgency mutation.
var createAgencyMutation = &graphql.Field{
	Name: "createAgency",
	Type: createAgencyPayloadType,
	Args: graphql.FieldConfigArgument{
		"input": &graphql.ArgumentConfig{
			Type: graphql.NewInputObject(createAgencyInputType),
		},
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		return func() (interface{}, error) {
			var payload createAgencyPayload
			input, err := toCreateAgencyInput(p.Args["input"].(map[string]interface{}))
			if err != nil {
				return payload, err
			}
			svc := p.Context.Value(service.ContextKeyAgencyService).(*service.AgencyService)
			agency, err := svc.CreateAgency(input.Name)
			if err != nil {
				return payload, err
			}
			payload.Agency = agency
			return payload, nil
		}, nil
	},
}
