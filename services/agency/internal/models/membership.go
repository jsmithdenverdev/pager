package models

import (
	"time"

	"github.com/jsmithdenverdev/pager/pkg/identity"
)

type MembershipStatus = string

const (
	MembershipStatusPending  MembershipStatus = "PENDING"
	MembershipStatusActive   MembershipStatus = "ACTIVE"
	MembershipStatusInactive MembershipStatus = "INACTIVE"
)

// membership represents a users membership in an agency including their role.
type Membership struct {
	PK         string           `dynamodbav:"pk"`
	SK         string           `dynamodbav:"sk"`
	Type       EntityType       `dynamodbav:"type"`
	Status     MembershipStatus `dynamodbav:"status"`
	Role       identity.Role    `dynamodbav:"role"`
	Created    time.Time        `dynamodbav:"created"`
	Modified   time.Time        `dynamodbav:"modified"`
	CreatedBy  string           `dynamodbav:"createdBy"`
	ModifiedBy string           `dynamodbav:"modifiedBy"`
}
