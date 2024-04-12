package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/jsmithdenverdev/pager/service"
)

// readDeviceQuery is the field definition for the device query.
var readDeviceQuery = &graphql.Field{
	Name: "device",
	Type: deviceType,
	Args: graphql.FieldConfigArgument{
		"id": &graphql.ArgumentConfig{
			Type: graphql.String,
		},
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		return func() (interface{}, error) {
			id := p.Args["id"].(string)
			svc := p.Context.Value(service.ContextKeyDeviceService).(*service.DeviceService)
			return svc.ReadDevice(id)
		}, nil
	},
}
