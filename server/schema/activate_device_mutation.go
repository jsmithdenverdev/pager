package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/jsmithdenverdev/pager/models"
	"github.com/jsmithdenverdev/pager/service"
)

// activateDeviceInput  represents the fields needed to activate a device.
type activateDeviceInput struct {
	Code     string `json:"code" validate:"required,min=1"`
	Endpoint string `json:"endpoint" validate:"required,url"`
}

// toActivateDeviceInput converts a `map[string]interface{}` into a
// `activateDeviceInput` performing validation on the model and returning any
// errors.
func toActivateDeviceInput(args map[string]interface{}) (activateDeviceInput, error) {
	var input activateDeviceInput
	code, ok := args["code"].(string)
	if !ok {
		code = ""
	}
	input.Code = code
	endpoint, ok := args["endpoint"].(string)
	if !ok {
		endpoint = ""
	}
	input.Endpoint = endpoint
	return input, validator.Struct(input)
}

// activateDeviceInputType  is the graphql input type for the activateDevice
// mutation.
var activateDeviceInputType = graphql.InputObjectConfig{
	Name: "ActivateDeviceInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"code": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
		"endpoint": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
	},
}

// activateDevicePayload is the struct representation of the result of a
// successful ActivateDevice mutation.
type activateDevicePayload struct {
	Device models.Device `json:"device"`
}

// activateDevicePayloadType  is the graphql representation of the result of a
// successful ActivateDevice mutation.
var activateDevicePayloadType = graphql.NewObject(graphql.ObjectConfig{
	Name: "ActivateDevicePayload",
	Fields: graphql.Fields{
		"device": &graphql.Field{
			Type: deviceType,
		},
	},
})

// activateDeviceMutation  is the field definition for the activateDevice
// mutation. Only the owner of a device can activate the device.
var activateDeviceMutation = &graphql.Field{
	Name: "activateDevice",
	Type: activateDevicePayloadType,
	Args: graphql.FieldConfigArgument{
		"input": &graphql.ArgumentConfig{
			Type: graphql.NewInputObject(activateDeviceInputType),
		},
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		return func() (interface{}, error) {
			var payload activateDevicePayload
			input, err := toActivateDeviceInput(p.Args["input"].(map[string]interface{}))
			if err != nil {
				return payload, err
			}
			svc := p.Context.Value(service.ContextKeyDeviceService).(*service.DeviceService)
			device, err := svc.ActivateDevice(input.Code, input.Endpoint)
			if err != nil {
				return payload, err
			}
			payload.Device = device
			return payload, nil
		}, nil
	},
}
