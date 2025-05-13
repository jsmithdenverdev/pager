package models

import (
	"time"

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
	PK         string           `dynamodbav:"pk"`
	SK         string           `dynamodbav:"sk"`
	Type       EntityType       `dynamodbav:"type"`
	Status     InvitationStatus `dynamodbav:"status"`
	Role       identity.Role    `dynamodbav:"role"`
	Created    time.Time        `dynamodbav:"created"`
	Modified   time.Time        `dynamodbav:"modified"`
	CreatedBy  string           `dynamodbav:"createdBy"`
	ModifiedBy string           `dynamodbav:"modifiedBy"`
}
