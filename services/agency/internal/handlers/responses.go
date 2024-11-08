package handlers

import "time"

type agencyResponse struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	Created    time.Time `json:"created"`
	CreatedBy  string    `json:"createdBy"`
	Modified   time.Time `json:"modified"`
	ModifiedBy string    `json:"modifedBy"`
	Address    string    `json:"address"`
	Contact    string    `json:"contact"`
}

type agenciesResponse struct {
	Records    []agencyResponse `json:"records"`
	NextCursor string           `json:"nextCursor"`
}

type membershipResponse struct {
	IDPID      string    `json:"idpid"`
	AgencyID   string    `json:"agencyId"`
	Created    time.Time `json:"created"`
	CreatedBy  string    `json:"createdBy"`
	Modified   time.Time `json:"modified"`
	ModifiedBy string    `json:"modifiedBy"`
}
