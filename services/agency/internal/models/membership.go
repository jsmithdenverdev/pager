package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/jsmithdenverdev/pager/pkg/dynarow"
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
	AgencyID          string           `dynamodbav:"agency_id"`
	UserID            string           `dynamodbav:"user_id"`
	Status            MembershipStatus `dynamodbav:"status"`
	Role              identity.Role    `dynamodbav:"role"`
	Created           time.Time        `dynamodbav:"created"`
	Modified          time.Time        `dynamodbav:"modified"`
	CreatedBy         string           `dynamodbav:"created_by"`
	ModifiedBy        string           `dynamodbav:"modified_by"`
	inverseMembership bool
}

func (a *Membership) Invert() {
	a.inverseMembership = true
}

func (a Membership) Type() string {
	return EntityTypeMembership
}

func (a Membership) EncodeKey() dynarow.Key {
	if a.inverseMembership {
		return dynarow.Key{
			PK: fmt.Sprintf("user#%s", a.UserID),
			SK: fmt.Sprintf("agency#%s", a.AgencyID),
		}
	}
	return dynarow.Key{
		PK: fmt.Sprintf("agency#%s", a.AgencyID),
		SK: fmt.Sprintf("user#%s", a.UserID),
	}
}

func (a Membership) DecodeKey(key dynarow.Key) error {
	pkParts := strings.Split(key.PK, "#")
	if len(pkParts) != 2 {
		return fmt.Errorf("invalid pk: %s", key.PK)
	}
	skParts := strings.Split(key.SK, "#")
	if len(skParts) != 2 {
		return fmt.Errorf("invalid sk: %s", key.PK)
	}
	if a.inverseMembership {
		a.AgencyID = skParts[1]
		a.UserID = pkParts[1]
	} else {
		a.AgencyID = pkParts[1]
		a.UserID = skParts[1]
	}

	return nil
}
