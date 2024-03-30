package main

type agencyStatus string

const (
	agencyStatusPending  agencyStatus = "PENDING"
	agencyStatusActive   agencyStatus = "ACTIVE"
	agencyStatusInactive agencyStatus = "INACTIVE"
)

// agency is the core entity of pager.
//
// An agency represents a real world agency (fire, police, ems, sar, etc.) that
// has a need to recieve pages via push notifications.
//
// Members of an agency are tracked as devices, to which notifications can be
// pushed.
type agency struct {
	auditable
	// ID is the UUID representing this agency in the pager system.
	ID string `json:"id" db:"id"`
	// Name is the name of the agency.
	Name   string       `json:"name" db:"name"`
	Status agencyStatus `json:"status" db:"status"`
}
