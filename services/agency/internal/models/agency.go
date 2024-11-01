package models

type AgencyStatus string

const (
	AgencyStatusPending  AgencyStatus = "PENDING"
	AgencyStatusActive   AgencyStatus = "ACTIVE"
	AgencyStatusInactive AgencyStatus = "INACTIVE"
)

// Agency is the core entity of pager.
//
// An Agency represents a real world agency (fire, police, ems, sar, etc.) that
// has a need to recieve pages via push notifications.
//
// Members of an Agency are tracked as Users, which have Devices to which
// notifications can be sent.
type Agency struct {
	Auditable
	ID string `dynamodbav:"agency_id"`
	// Name is the name of the agency.
	Name   string       `dynamodbav:"name"`
	Type   string       `dynamodbav:"type"`
	Status AgencyStatus `dynamodbav:"status"`
}

// Member represents an inverse relationship between an agency and a
// user. This allows us to query users for an agency.
type Member struct {
	Auditable
	AgencyID string `dynamodbav:"agency_id"`
	IDPID    string `dynamodbav:"idpid"`
	Type     string `dynamodbav:"type"`
}
