package schema

import (
	"errors"

	"github.com/graphql-go/graphql"
	"github.com/jsmithdenverdev/pager/models"
	"github.com/jsmithdenverdev/pager/service"
)

// inviteUserInput represents the fields needed to invite a user.
type inviteUserInput struct {
	Email     string      `json:"email" validate:"min=1,email"`
	AccountID string      `json:"agencyId" validate:"required,uuid"`
	Role      models.Role `json:"role" validate:"required"`
}

// inviteUserInputType  is the graphql input type for the inviteUser
// mutation.
var inviteUserInputType = graphql.InputObjectConfig{
	Name: "InviteUserInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"email": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"agencyId": &graphql.InputObjectFieldConfig{
			Type: graphql.ID,
		},
		"role": &graphql.InputObjectFieldConfig{
			Type: roleType,
		},
	},
}

// toCreateAgencyInput converts a `map[string]interface{}` into a
// `createAgencyInput` performing validation on the model and returning any
// errors.
func toInviteUserInput(args map[string]interface{}) (inviteUserInput, error) {
	var input inviteUserInput
	email, ok := args["email"].(string)
	if !ok {
		email = ""
	}
	input.Email = email
	agencyId, ok := args["agencyId"].(string)
	if !ok {
		agencyId = ""
	}
	input.AccountID = agencyId
	role, ok := args["role"].(models.Role)
	if !ok {
		role = ""
	}
	input.Role = role
	return input, validator.Struct(input)
}

// inviteUserPayload is the struct representation of the result of a successful
// inviteUser mutation.
type inviteUserPayload struct {
	User models.User `json:"user"`
}

// inviteUserPayloadType is the graphql representation of the result of a
// successful inviteUser mutation.
var inviteUserPayloadType = graphql.NewObject(graphql.ObjectConfig{
	Name: "InviteUserPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: userType,
		},
	},
})

// inviteUserMutation is the field definition for the inviteUser mutation.
var inviteUserMutation = &graphql.Field{
	Name: "inviteUser",
	Type: inviteUserPayloadType,
	Args: graphql.FieldConfigArgument{
		"input": &graphql.ArgumentConfig{
			Type: graphql.NewInputObject(inviteUserInputType),
		},
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		return func() (interface{}, error) {
			var payload inviteUserPayload
			input, err := toInviteUserInput(p.Args["input"].(map[string]interface{}))
			if err != nil {
				return payload, err
			}
			if input.Role == models.RolePlatformAdmin {
				return payload, errors.New("cannot invite platform_admins")
			}
			svc := p.Context.Value(service.ContextKeyAgencyService).(*service.AgencyService)
			user, err := svc.InviteUser(input.Email, input.AccountID, input.Role)
			if err != nil {
				return payload, err
			}
			payload.User = user
			return payload, nil
		}, nil
	},
}
