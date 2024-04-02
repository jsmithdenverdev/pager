package models

import "time"

type Auditable struct {
	Created    time.Time `json:"created" db:"created"`
	CreatedBy  string    `json:"createdBy" db:"created_by"`
	Modified   time.Time `json:"modified" db:"modified"`
	ModifiedBy string    `json:"modifiedBy" db:"modified_by"`
}
