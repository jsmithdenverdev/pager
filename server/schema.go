package main

import (
	"fmt"
	"log/slog"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/go-playground/validator/v10"
	"github.com/graph-gophers/dataloader/v7"
	"github.com/graphql-go/graphql"
	"github.com/jmoiron/sqlx"
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
func newSchema(
	config config,
	logger *slog.Logger,
	validate *validator.Validate,
	authz *authzed.Client,
	db *sqlx.DB) (graphql.Schema, error) {
	schemaConfig := graphql.SchemaConfig{}

	types := newGraphTypes(config, logger, validate, authz, db)
	registerQueries(&schemaConfig, config, types, logger, validate, authz, db)
	registerMutations(&schemaConfig, config, types, logger, validate, authz, db)

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
func registerQueries(
	schema *graphql.SchemaConfig,
	config config,
	types graphTypes,
	logger *slog.Logger,
	validate *validator.Validate,
	authz *authzed.Client,
	db *sqlx.DB) {
	// Register additional graphql queries here
	queries := []*graphql.Field{
		readAgencyQuery(logger, types),
		userInfoQuery(logger, types, authz, db),
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
func registerMutations(
	schema *graphql.SchemaConfig,
	config config,
	types graphTypes,
	logger *slog.Logger,
	validate *validator.Validate,
	authz *authzed.Client,
	db *sqlx.DB) {
	// Register additional graphql mutations here
	mutations := []*graphql.Field{
		createAgencyMutation(logger, types, validate, authz, db),
		updateAgencyMutation(logger),
		deleteAgencyMutation(logger),
	}
	var rootMutation = graphql.ObjectConfig{
		Name:   "RootMutation",
		Fields: toFields(mutations),
	}
	schema.Mutation = graphql.NewObject(rootMutation)
}

// graphTypes holds references to shared GraphQL types. GraphQL types must have
// unique names within a schema, which requires the types being initialized
// exactly once. This struct can be passed as an argument to query and mutation
// constructors, allowing them to reference the shared graph types.
type graphTypes struct {
	agency *graphql.Object
	user   *graphql.Object
}

// newGraphTypes returns an initialized graphTypes object.
func newGraphTypes(config config,
	logger *slog.Logger,
	validate *validator.Validate,
	authz *authzed.Client,
	db *sqlx.DB) graphTypes {
	// Invoke type generators here. Types need to be created in the order they
	// are used in subtypes. E.g., if a type has a dependency on another type
	// (e.g., for a field) the sub type must be created first.
	agencyType := agencyType()
	userType := userType(logger, agencyType, authz, db)

	return graphTypes{
		agency: agencyType,
		user:   userType,
	}
}

// dataLoaders holds references to request scoped data loaders. newDataLoaders
// is called from a piece of middleware and initializes a dataLoaders for the
// lifetime of that request. All read and list operations should be done via a
// dataloader.

type dataLoaders struct {
	checkPermission    *dataloader.Loader[*v1.CheckPermissionRequest, *v1.CheckPermissionResponse]
	readAgency         *dataloader.Loader[string, agency]
	listAgenciesByUser *dataloader.Loader[string, []agency]
}

func newDataLoaders(config config,
	logger *slog.Logger,
	validate *validator.Validate,
	authz *authzed.Client,
	db *sqlx.DB) dataLoaders {
	return dataLoaders{
		checkPermission:    newCheckPermissionDataLoader(authz),
		readAgency:         newReadAgencyDataLoader(db),
		listAgenciesByUser: newListAgenciesByUserDataloader(logger, db),
	}
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

// toResultType wraps a `*graphql.Object` with a union  type. This union has a
// `Result` member for the original object, as well as a member for each type of
// provided error.
func toResultType[Result any](resultType *graphql.Object, errTypes ...*graphql.Object) *graphql.Union {
	var types []*graphql.Object
	types = append(types, resultType)
	types = append(types, errTypes...)
	return graphql.NewUnion(graphql.UnionConfig{
		Name:  fmt.Sprintf("%sResult", resultType.Name()),
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
