package schema

import "github.com/graphql-go/graphql"

func New() (graphql.Schema, error) {
	schemaConfig := graphql.SchemaConfig{}
	registerQueries(&schemaConfig)
	registerMutations(&schemaConfig)
	return graphql.NewSchema(schemaConfig)
}

func registerQueries(schema *graphql.SchemaConfig) {
	// Register queries here
	queries := []*graphql.Field{
		readAgencyQuery,
		listAgenciesQuery,
		userInfoQuery,
	}
	var rootQuery = graphql.ObjectConfig{
		Name:   "Query",
		Fields: toFields(queries),
	}
	schema.Query = graphql.NewObject(rootQuery)
}

func registerMutations(schema *graphql.SchemaConfig) {
	// Register mutations here
	mutations := []*graphql.Field{
		createAgencyMutation,
	}
	var rootMutation = graphql.ObjectConfig{
		Name:   "Mutation",
		Fields: toFields(mutations),
	}
	schema.Mutation = graphql.NewObject(rootMutation)
}

// toFields converts a slice of `*graphql.Field` into a `graphql.Fields`. All
// fields in the slice must have a `Name`, which is used as the key in
// `graphql.Fields`.
func toFields(fields []*graphql.Field) graphql.Fields {
	var f graphql.Fields = make(graphql.Fields)
	for _, field := range fields {
		f[field.Name] = field
	}
	return f
}
