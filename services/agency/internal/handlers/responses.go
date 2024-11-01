package handlers

import "time"

// createAgencyRequest is the data returned on successful creation of an agency.
type agencyResponse struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	Created    time.Time `json:"created"`
	CreatedBy  string    `json:"createdBy"`
	Modified   time.Time `json:"modified"`
	ModifiedBy string    `json:"modifedBy"`
}

type agenciesResponse struct {
	Records []struct {
		ID string `json:"id"`
	} `json:"records"`
	Pagination struct {
	} `json:"pagination"`
}
