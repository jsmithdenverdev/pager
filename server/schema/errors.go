package schema

import (
	goplaygroundvalidator "github.com/go-playground/validator/v10"
	"github.com/graphql-go/graphql"
	"github.com/jsmithdenverdev/pager/authz"
)

// errorType is a `*grahpql.Interface` named `Error` that describes error
// fields.
var errorType = graphql.NewInterface(graphql.InterfaceConfig{
	Name: "Error",
	Fields: graphql.Fields{
		"error": &graphql.Field{
			Type: graphql.String,
		},
	},
})

// baseErrorType is a `*graphql.Object` implementing the `Error` interface. It
// represents the most generic form of error that the API can return.
var baseErrorType = graphql.NewObject(graphql.ObjectConfig{
	Name: "BaseError",
	Interfaces: []*graphql.Interface{
		errorType,
	},
	Fields: graphql.Fields{
		"error": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(error).Error(), nil
			},
		},
	},
	IsTypeOf: func(p graphql.IsTypeOfParams) bool {
		_, ok := p.Value.(error)
		return ok
	},
})

// authzErrorType is a `*graphql.Object` implementing the `Error` interface. It
// represents authorization errors.
var authzErrorType = graphql.NewObject(graphql.ObjectConfig{
	Name: "AuthorizationError",
	Interfaces: []*graphql.Interface{
		errorType,
	},
	Fields: graphql.Fields{
		"error": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(authz.AuthzError).Error(), nil
			},
		},
	},
	IsTypeOf: func(p graphql.IsTypeOfParams) bool {
		_, ok := p.Value.(authz.AuthzError)
		return ok
	},
})

// validationErrorType is a `*graphql.Object` implementing the `Error`
// interface. It represents the graphql version of `validator.ValidationErrors`.
var validationErrorType = graphql.NewObject(graphql.ObjectConfig{
	Name: "ValidationError",
	Interfaces: []*graphql.Interface{
		errorType,
	},
	Fields: graphql.Fields{
		"error": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(goplaygroundvalidator.ValidationErrors).Error(), nil
			},
		},
		"fields": &graphql.Field{
			Name: "ValidationErrorFields",
			Type: graphql.NewList(graphql.NewObject(graphql.ObjectConfig{
				Name: "ValidationErrorField",
				Interfaces: []*graphql.Interface{
					errorType,
				},
				Fields: graphql.Fields{
					"error": &graphql.Field{
						Type: graphql.String,
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							return p.Source.(goplaygroundvalidator.FieldError).Error(), nil
						},
					},
					"field": &graphql.Field{
						Type: graphql.String,
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							return p.Source.(goplaygroundvalidator.FieldError).StructField(), nil
						},
					},
					"value": &graphql.Field{
						Type: graphql.String,
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							return p.Source.(goplaygroundvalidator.FieldError).Value(), nil
						},
					},
					"type": &graphql.Field{
						Type: graphql.String,
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							return p.Source.(goplaygroundvalidator.FieldError).Type().String(), nil
						},
					},
				},
			})),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(goplaygroundvalidator.ValidationErrors), nil
			},
		},
	},
	IsTypeOf: func(p graphql.IsTypeOfParams) bool {
		_, ok := p.Value.(goplaygroundvalidator.ValidationErrors)
		return ok
	},
})
