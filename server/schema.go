package main

import (
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/graphql-go/graphql"
)

// newSchema creates a new `graphql.Schema` used to describe our application.
// All queries and mutations are exposed through this schema as fields.
//
// The application has `RootQuery` and `RootMutation` fields, exposing queries
// and mutations respectively. These fields are all registered from
// `registerQueries` and `registerMutations`.
//
// Fields are typically represented using a function that accept the
// dependencies for the field as arguments, and return a `*graphql.Field`.
//
// newSchema accepts the full set of dependencies for all of its fields as
// arguments.
func newSchema(logger *slog.Logger, validate *validator.Validate) (graphql.Schema, error) {
	schemaConfig := graphql.SchemaConfig{}

	registerQueries(&schemaConfig, logger, validate)
	registerMutations(&schemaConfig, logger, validate)

	if schema, err := graphql.NewSchema(schemaConfig); err != nil {
		return graphql.Schema{}, err
	} else {
		return schema, nil
	}
}

// registerQueries adds a `RootQuery` field to the provided schema. Query fields
// are then registered on `RootQuery`.
//
// Fields are typically represented using a function that accept the
// dependencies for the field as arguments, and return a `*graphql.Field`.
//
// registerQueries accepts the full set of dependencies for all of its queries
// as arguments.
func registerQueries(schema *graphql.SchemaConfig, logger *slog.Logger, validate *validator.Validate) {
	// Register additional graphql queries here
	queries := []*graphql.Field{
		readAgencyQuery(logger),
	}
	var rootQuery = graphql.ObjectConfig{
		Name:   "RootQuery",
		Fields: toFields(queries),
	}
	schema.Query = graphql.NewObject(rootQuery)
}

// registerMutations adds a `RootMutation` field to the provided schema.
// Mutation fields are then registered on `RootMutation`.
//
// Fields are typically represented using a function that accept the
// dependencies for the field as arguments, and return a `*graphql.Field`.
//
// registerMutations accepts the full set of dependencies for all of its queries
// as arguments.
func registerMutations(schema *graphql.SchemaConfig, logger *slog.Logger, validate *validator.Validate) {
	// Register additional graphql mutations here
	mutations := []*graphql.Field{
		createAgencyMutation(logger, validate),
		updateAgencyMutation(logger),
		deleteAgencyMutation(logger),
	}
	var rootMutation = graphql.ObjectConfig{
		Name:   "RootMutation",
		Fields: toFields(mutations),
	}
	schema.Mutation = graphql.NewObject(rootMutation)
}

// toFields converts a slice of `*graphql.Field` into a `graphql.Fields` Fields
// in the slice must have a `Name`, which is used as the key in
// `graphql.Fields`.
func toFields(fields []*graphql.Field) graphql.Fields {
	var f graphql.Fields = make(graphql.Fields)
	for _, field := range fields {
		f[field.Name] = field
	}
	return f
}

// newResultType wraps a `*graphql.Object` with a union  type. This union has a
// `Result` member for the original object, as well as a member for each type of
// provided error.
func newResultType[Result any](name string, resultType *graphql.Object, errTypes ...*graphql.Object) *graphql.Union {
	var types []*graphql.Object
	types = append(types, resultType)
	types = append(types, errTypes...)
	return graphql.NewUnion(graphql.UnionConfig{
		Name:  name,
		Types: types,
		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
			// To check the type, we need to go from most specific to least specific
			// the last item in this list should be the error interface.
			if _, ok := p.Value.(Result); ok {
				return resultType
			}
			if _, ok := p.Value.(authzError); ok {
				return authzErrorType
			}
			if _, ok := p.Value.(validator.ValidationErrors); ok {
				return validationErrorType
			}
			if _, ok := p.Value.(error); ok {
				return baseErrorType
			}
			return nil
		},
	})
}
