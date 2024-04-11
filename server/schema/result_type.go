package schema

import (
	"fmt"

	goplaygroundvalidator "github.com/go-playground/validator/v10"
	"github.com/graphql-go/graphql"
	"github.com/jsmithdenverdev/pager/authz"
)

// toResultType wraps a `*graphql.Object` with a union  type. This union has a
// `Result` member for the original object, as well as a member for each type of
// provided error.
func toResultType[Result any](resultType *graphql.Object, errTypes ...*graphql.Object) *graphql.Union {
	panic("deprecated")
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
			if _, ok := p.Value.(authz.AuthzError); ok {
				return authzErrorType
			}
			if _, ok := p.Value.(goplaygroundvalidator.ValidationErrors); ok {
				return validationErrorType
			}
			if _, ok := p.Value.(error); ok {
				return baseErrorType
			}
			return nil
		},
	})
}

// toResultType wraps a `*graphql.Object` with a union  type. This union has a
// `Result` member for the original object, as well as a member for each type of
// provided error.
func toListResultType[Result any](resultType *graphql.Object, errTypes ...*graphql.Object) *graphql.Union {
	panic("deprecated")
	var types []any
	types = append(types, graphql.NewList(resultType))
	for _, errType := range errTypes {
		types = append(types, errType)
	}

	var unionTypes []*graphql.Object
	for _, t := range types {
		unionTypes = append(unionTypes, t.(*graphql.Object))
	}
	return graphql.NewUnion(graphql.UnionConfig{
		Name:  fmt.Sprintf("List%sResult", resultType.Name()),
		Types: unionTypes,
		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
			// To check the type, we need to go from most specific to least specific
			// the last item in this list should be the error interface.
			if _, ok := p.Value.(Result); ok {
				return resultType
			}
			if _, ok := p.Value.(authz.AuthzError); ok {
				return authzErrorType
			}
			if _, ok := p.Value.(goplaygroundvalidator.ValidationErrors); ok {
				return validationErrorType
			}
			if _, ok := p.Value.(error); ok {
				return baseErrorType
			}
			return nil
		},
	})
}
