package models

import "github.com/jsmithdenverdev/pager/pkg/identity"

type User struct {
	Keys
	Auditable
	Name         string                   `dynamodbav:"name"`
	Email        string                   `dynamodbav:"email"`
	Entitlements []identity.Entitlement   `dynamodbav:"entitlements"`
	Memberships  map[string]identity.Role `dynamodbav:"memberships"`
}
