package authz

import (
	"context"
)

type Resource struct {
	Type string
	ID   string
}

type Action struct {
	Type string
	ID   string
}

type Authorizer interface {
	IsAuthorized(context.Context, Resource, Action) (bool, error)
}
