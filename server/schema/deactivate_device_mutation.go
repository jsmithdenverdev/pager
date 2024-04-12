package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/jsmithdenverdev/pager/models"
	"github.com/jsmithdenverdev/pager/service"
)

// deactivateDeviceInput  represents the fields needed to deactivate a device.
type deactivateDeviceInput struct {
	ID string `json:"id" validate:"required,uuid"`
}

// toDeactivateDeviceInput converts a `map[string]interface{}` into a
// `deactivateDeviceInput` performing validation on the model and returning any
// errors.
func toDeactivateDeviceInput(args map[string]interface{}) (deactivateDeviceInput, error) {
	var input deactivateDeviceInput
	id, ok := args["id"].(string)
	if !ok {
		id = ""
	}
	input.ID = id
	return input, validator.Struct(input)
}

// deactivateDeviceInputType  is the graphql input type for the deactivateDevice
// mutation.
var deactivateDeviceInputType = graphql.InputObjectConfig{
	Name: "DeactivateDeviceInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"id": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.ID),
		},
	},
}

// deactivateDevicePayload is the struct representation of the result of a
// successful DeactivateDevice mutation.
type deactivateDevicePayload struct {
	Device models.Device `json:"device"`
}

// deactivateDevicePayloadType  is the graphql representation of the result of a
// successful DeactivateDevice mutation.
var deactivateDevicePayloadType = graphql.NewObject(graphql.ObjectConfig{
	Name: "DeactivateDevicePayload",
	Fields: graphql.Fields{
		"device": &graphql.Field{
			Type: deviceType,
		},
	},
})

// deactivateDeviceMutation  is the field definition for the deactivateDevice
// mutation. Only the owner of a device can deactivate the device.
var deactivateDeviceMutation = &graphql.Field{
	Name: "deactivateDevice",
	Type: deactivateDevicePayloadType,
	Args: graphql.FieldConfigArgument{
		"input": &graphql.ArgumentConfig{
			Type: graphql.NewInputObject(deactivateDeviceInputType),
		},
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		return func() (interface{}, error) {
			var payload deactivateDevicePayload
			input, err := toDeactivateDeviceInput(p.Args["input"].(map[string]interface{}))
			if err != nil {
				return payload, err
			}
			svc := p.Context.Value(service.ContextKeyDeviceService).(*service.DeviceService)
			device, err := svc.DeactivateDevice(input.ID)
			if err != nil {
				return payload, err
			}
			payload.Device = device
			return payload, nil
		}, nil
	},
}
