package app

import (
	"strings"
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

// lookup is a record used to lookup a user by various attributes. It consists
// of a pk which is the type of lookup (e.g. email), and a sk which is the users
// id.
type lookup struct {
	PK   string     `dynamodbav:"pk"`
	SK   string     `dynamodbav:"sk"`
	Type entityType `dynamodbav:"type"`
}

func (l *lookup) UserID() string {
	return strings.Split(l.SK, "#")[1]
}
