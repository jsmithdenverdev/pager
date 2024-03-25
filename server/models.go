package main

import "time"

// auditable represents an auditable entity in our system.
type auditable struct {
	Created    time.Time `json:"created" db:"created"`
	CreatedBy  string    `json:"createdBy" db:"created_by"`
	Modified   time.Time `json:"modified" db:"modified"`
	ModifiedBy string    `json:"modifiedBy" db:"modified_by"`
}
