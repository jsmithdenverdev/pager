package app

type agencyStatus = string

const (
	agencyStatusActive   = "ACTIVE"
	agencyStatusInactive = "INACTIVE"
)

type membershipStatus = string

const (
	membershipStatusPending  = "PENDING"
	membershipStatusActive   = "ACTIVE"
	membershipStatusInactive = "INACTIVE"
)

type invitationStatus = string

const (
	invitationStatusPending  = "PENDING"
	invitationStatusComplete = "COMPLETE"
	invitationStatusDeclined = "DECLINED"
	invitationStatusExpired  = "EXPIRED"
)

type registrationStatus = string

const (
	registrationStatusPending  = "PENDING"
	registrationStatusComplete = "COMPLETE"
	registrationStatusDeclined = "DECLINED"
	registrationStatusExpired  = "EXPIRED"
)
