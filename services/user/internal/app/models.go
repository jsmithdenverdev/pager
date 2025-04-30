package app

import (
	"time"

	"github.com/jsmithdenverdev/pager/pkg/identity"
)

type user struct {
	PK           string                   `dynamodbav:"pk"`
	SK           string                   `dynamodbav:"sk"`
	Name         string                   `dynamodbav:"name"`
	Type         entityType               `dynamodbav:"type"`
	Email        string                   `dynamodbav:"email"`
	Entitlements []identity.Entitlement   `dynamodbav:"entitlements"`
	Memberships  map[string]identity.Role `dynamodbav:"memberships"`
	Created      time.Time                `dynamodbav:"created"`
	Modified     time.Time                `dynamodbav:"modified"`
	CreatedBy    string                   `dynamodbav:"createdBy"`
	ModifiedBy   string                   `dynamodbav:"modifiedBy"`
}

type userLookup struct {
	PK   string     `dynamodbav:"pk"`
	SK   string     `dynamodbav:"sk"`
	Type entityType `dynamodbav:"type"`
}
