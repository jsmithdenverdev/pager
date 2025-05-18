package models

type EntityType = string

const (
	EntityTypeAgency       EntityType = "AGENCY"
	EntityTypeMembership   EntityType = "MEMBERSHIP"
	EntityTypeInvitation   EntityType = "INVITATION"
	EntityTypeRegistration EntityType = "REGISTRATION"
)
