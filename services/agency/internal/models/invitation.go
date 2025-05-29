package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/jsmithdenverdev/pager/pkg/dynarow"

	"github.com/jsmithdenverdev/pager/pkg/identity"
)

type InvitationStatus = string

const (
	InvitationStatusPending  InvitationStatus = "PENDING"
	InvitationStatusComplete InvitationStatus = "COMPLETE"
	InvitationStatusDeclined InvitationStatus = "DECLINED"
	InvitationStatusExpired  InvitationStatus = "EXPIRED"
)

// Invitation represents an invitation to join an agency.
type Invitation struct {
	AgencyID   string           `dynamodbav:"agency_id"`
	Email      string           `dynamodbav:"email"`
	Status     InvitationStatus `dynamodbav:"status"`
	Role       identity.Role    `dynamodbav:"role"`
	Created    time.Time        `dynamodbav:"created"`
	Modified   time.Time        `dynamodbav:"modified"`
	CreatedBy  string           `dynamodbav:"createdBy"`
	ModifiedBy string           `dynamodbav:"modifiedBy"`
}

func (a *Invitation) Type() string {
	return EntityTypeInvitation
}

func (a *Invitation) EncodeKey() dynarow.Key {
	return dynarow.Key{
		PK: fmt.Sprintf("invite#%s", a.Email),
		SK: fmt.Sprintf("agency#%s", a.AgencyID),
	}
}

func (a *Invitation) DecodeKey(key dynarow.Key) error {
	pkParts := strings.Split(key.PK, "#")
	if len(pkParts) != 2 {
		return fmt.Errorf("invalid pk: %s", key.PK)
	}
	skParts := strings.Split(key.SK, "#")
	if len(skParts) != 2 {
		return fmt.Errorf("invalid sk: %s", key.PK)
	}
	a.Email = pkParts[1]
	a.CreatedBy = skParts[1]
	return nil
}
