package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/jsmithdenverdev/pager/service"
)

// userInfoQuery is the field definition for the userInfo query.
var userInfoQuery = &graphql.Field{
	Name: "userInfo",
	Type: userType,
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		return func() (interface{}, error) {
			userService := p.Context.Value(service.ContextKeyUserService).(*service.UserService)
			return userService.Info()
		}, nil
	},
}
