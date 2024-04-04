package authz

import "fmt"

type AuthzError struct {
	Permission permission
	Resource   Resource
}

func (p AuthzError) Error() string {
	return fmt.Sprintf("actor is not authorized to %s on %s %s", p.Permission, p.Resource.Type, p.Resource.ID)
}

func NewAuthzError(permission permission, resource Resource) AuthzError {
	return AuthzError{
		Permission: permission,
		Resource:   resource,
	}
}
