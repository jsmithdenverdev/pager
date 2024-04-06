package schema

import (
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	jwtvalidator "github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/graphql-go/graphql"
	"github.com/jsmithdenverdev/pager/models"
	"github.com/jsmithdenverdev/pager/service"
)

// createAgencyInput represents the fields needed to create a new agency.
type createAgencyInput struct {
	Name string `json:"name" validate:""`
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

var createAgencyMutation = &graphql.Field{
	Name: "createAgency",
	Type: toResultType[createAgencyPayload](
		createAgencyPayloadType,
		baseErrorType,
		validationErrorType,
		authzErrorType),
	Args: graphql.FieldConfigArgument{
		"input": &graphql.ArgumentConfig{
			Type: graphql.NewInputObject(createAgencyInputType),
		},
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		var payload createAgencyPayload
		input, err := toCreateAgencyInput(p.Args)
		if err != nil {
			return payload, err
		}
		claims := p.Context.Value(jwtmiddleware.ContextKey{}).(*jwtvalidator.ValidatedClaims)
		svc := p.Context.Value(service.ContextKeyAgencyService).(*service.AgencyService)
		agency, err := svc.Create(input.Name, claims.RegisteredClaims.Subject)
		if err != nil {
			return payload, err
		}
		payload.Agency = agency
		return payload, nil
	},
}
