package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/jsmithdenverdev/pager/models"
)

// roleType is a graphql enum representing the role of a user.
var roleType = graphql.NewEnum(graphql.EnumConfig{
	Name: "Role",
	Values: graphql.EnumValueConfigMap{
		"READER": &graphql.EnumValueConfig{
			Value: models.RoleReader,
		},
		"WRITER": &graphql.EnumValueConfig{
			Value: models.RoleWriter,
		},
		"PLATFORM_ADMIN": &graphql.EnumValueConfig{
			Value: models.RolePlatformAdmin,
		},
	},
})

// userRoleType is the object definition for a user role.
var userRoleType = graphql.NewObject(graphql.ObjectConfig{
	Name: "UserRole",
	Fields: graphql.Fields{
		"role": &graphql.Field{
			Type: roleType,
		},
		"userId": &graphql.Field{
			Type: graphql.ID,
		},
		"agencyId": &graphql.Field{
			Type: graphql.ID,
		},
	},
})
