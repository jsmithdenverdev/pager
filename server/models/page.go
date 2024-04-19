package models

// Page is a core entity of pager.
//
// A page represents a notification being distrubuted to devices.
type Page struct {
	Auditable
	AgencyID string `json:"agencyId" db:"agency_id"`
	Title    string `json:"title" db:"title"`
	Content  string `json:"content" db:"content"`
}
