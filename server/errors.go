package main

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/graphql-go/graphql"
)

// authzError is an authorization error for a particular Actor, Resource, and
// Action.
//
// authzError implementes the error interface.
type authzError struct {
	Actor    string `json:"actor"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// newAuthzError creates a new `authzError` for the provided actor, resource,
// and action.
func newAuthzError(actor, resource, action string) authzError {
	return authzError{
		Actor:    actor,
		Resource: resource,
		Action:   action,
	}
}

// Error returns an error string for this `authzError`.
func (err authzError) Error() string {
	return fmt.Sprintf("%s is not permitted to %s on %s", err.Actor, err.Action, err.Resource)
}

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
				return p.Source.(authzError).Error(), nil
			},
		},
	},
	IsTypeOf: func(p graphql.IsTypeOfParams) bool {
		_, ok := p.Value.(authzError)
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
				return p.Source.(validator.ValidationErrors).Error(), nil
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
							return p.Source.(validator.FieldError).Error(), nil
						},
					},
					"field": &graphql.Field{
						Type: graphql.String,
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							return p.Source.(validator.FieldError).StructField(), nil
						},
					},
					"value": &graphql.Field{
						Type: graphql.String,
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							return p.Source.(validator.FieldError).Value(), nil
						},
					},
					"type": &graphql.Field{
						Type: graphql.String,
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							return p.Source.(validator.FieldError).Type().String(), nil
						},
					},
				},
			})),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(validator.ValidationErrors), nil
			},
		},
	},
	IsTypeOf: func(p graphql.IsTypeOfParams) bool {
		_, ok := p.Value.(validator.ValidationErrors)
		return ok
	},
})
