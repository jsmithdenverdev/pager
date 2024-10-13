package authz

import "fmt"

type UnauthorizedError struct {
	Resource Resource
	Action   Action
}

func (err UnauthorizedError) Error() string {
	return fmt.Sprintf(
		"user is not authorized to perform action %s on resource %s",
		fmt.Sprintf("%s::%s", err.Action.Type, err.Action.ID),
		fmt.Sprintf("%s::%s", err.Resource.Type, err.Resource.ID),
	)
}

func NewUnauthorizedError(resource Resource, action Action) UnauthorizedError {
	return UnauthorizedError{
		Resource: resource,
		Action:   action,
	}
}
