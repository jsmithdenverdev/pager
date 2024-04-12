package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/jsmithdenverdev/pager/models"
	"github.com/jsmithdenverdev/pager/service"
)

// provisionDeviceInput  represents the fields needed to provision a device.
type provisionDeviceInput struct {
	Name     string `json:"name" validate:"required,min=1"`
	AgencyID string `json:"agencyId" validate:"required,uuid"`
	OwnerID  string `json:"ownerId" validate:"required,uuid"`
}

// toProvisionDeviceInput converts a `map[string]interface{}` into a
// `provisionDeviceInput` performing validation on the model and returning any
// errors.
func toProvisionDeviceInput(args map[string]interface{}) (provisionDeviceInput, error) {
	var input provisionDeviceInput
	name, ok := args["name"].(string)
	if !ok {
		name = ""
	}
	input.Name = name
	agencyId, ok := args["agencyId"].(string)
	if !ok {
		agencyId = ""
	}
	input.AgencyID = agencyId
	ownerId, ok := args["ownerId"].(string)
	if !ok {
		ownerId = ""
	}
	input.OwnerID = ownerId
	return input, validator.Struct(input)
}

// provisionDeviceInputType  is the graphql input type for the provisionDevice
// mutation.
var provisionDeviceInputType = graphql.InputObjectConfig{
	Name: "ProvisionDeviceInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"name": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"agencyId": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.ID),
		},
		"ownerId": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "The pager user ID for the user who owns this device.",
		},
	},
}

// provisionDevicePayload is the struct representation of the result of a
// successful ProvisionDevice mutation.
type provisionDevicePayload struct {
	Device models.Device `json:"device"`
}

// provisionDevicePayloadType  is the graphql representation of the result of a
// successful ProvisionDevice mutation.
var provisionDevicePayloadType = graphql.NewObject(graphql.ObjectConfig{
	Name: "ProvisionDevicePayload",
	Fields: graphql.Fields{
		"device": &graphql.Field{
			Type: deviceType,
		},
	},
})

// provisionDeviceMutation  is the field definition for the provisionDevice
// mutation.
var provisionDeviceMutation = &graphql.Field{
	Name: "provisionDevice",
	Type: provisionDevicePayloadType,
	Args: graphql.FieldConfigArgument{
		"input": &graphql.ArgumentConfig{
			Type: graphql.NewInputObject(provisionDeviceInputType),
		},
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		return func() (interface{}, error) {
			var payload provisionDevicePayload
			input, err := toProvisionDeviceInput(p.Args["input"].(map[string]interface{}))
			if err != nil {
				return payload, err
			}
			svc := p.Context.Value(service.ContextKeyDeviceService).(*service.DeviceService)
			device, err := svc.ProvisionDevice(input.AgencyID, input.OwnerID, input.Name)
			if err != nil {
				return payload, err
			}
			payload.Device = device
			return payload, nil
		}, nil
	},
}
