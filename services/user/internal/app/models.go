package app

import (
	"time"

	"github.com/jsmithdenverdev/pager/pkg/identity"
)

//-----------------------------------------------------------------------------
// FIELDS
//-----------------------------------------------------------------------------

type keyFields struct {
	PK   string     `dynamodbav:"pk"`
	SK   string     `dynamodbav:"sk"`
	Type entityType `dynamodbav:"type"`
}

type auditableFields struct {
	Created    time.Time `dynamodbav:"created"`
	Modified   time.Time `dynamodbav:"modified"`
	CreatedBy  string    `dynamodbav:"createdBy"`
	ModifiedBy string    `dynamodbav:"modifiedBy"`
}

func newAuditableFields(userID string, timestamp time.Time) auditableFields {
	return auditableFields{
		Created:    timestamp,
		Modified:   timestamp,
		CreatedBy:  userID,
		ModifiedBy: userID,
	}
}

type user struct {
	keyFields
	auditableFields
	Name         string                   `dynamodbav:"name"`
	Email        string                   `dynamodbav:"email"`
	Entitlements []identity.Entitlement   `dynamodbav:"entitlements"`
	Memberships  map[string]identity.Role `dynamodbav:"memberships"`
}

// lookup is a record used to lookup a user by various attributes. It consists
// of a pk which is the type of lookup (e.g. email) and the users id.
type lookup struct {
	keyFields
	auditableFields
	UserID string `dynamodbav:"userId"`
}
