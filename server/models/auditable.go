package models

import "time"

type Auditable struct {
	ID         string    `json:"id" db:"id"`
	Created    time.Time `json:"created" db:"created"`
	CreatedBy  string    `json:"createdBy" db:"created_by"`
	Modified   time.Time `json:"modified" db:"modified"`
	ModifiedBy string    `json:"modifiedBy" db:"modified_by"`
}

func (auditable Auditable) Identity() string {
	return auditable.ID
}
