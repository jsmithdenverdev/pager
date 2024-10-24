package authz

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions/types"
)

// UnauthorizedError represents an authorization failure for a given Entity
// and Action.
type UnauthorizedError struct {
	Entity *types.EntityIdentifier
	Action *types.ActionIdentifier
}

// Error implements the error interface.
func (err UnauthorizedError) Error() string {
	return fmt.Sprintf(
		"user is not authorized to perform action %s on resource %s",
		fmt.Sprintf("%s::%s", *err.Action.ActionType, *err.Action.ActionId),
		fmt.Sprintf("%s::%s", *err.Entity.EntityType, *err.Entity.EntityId),
	)
}

// NewUnauthorizedError returns a new instance of an UnauthorizedError.
func NewUnauthorizedError(resource *types.EntityIdentifier, action *types.ActionIdentifier) UnauthorizedError {
	return UnauthorizedError{
		Entity: resource,
		Action: action,
	}
}
