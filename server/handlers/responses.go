package handlers

import "github.com/jsmithdenverdev/pager/models"

// agencyResponse is the response model representation of an Agency.
type agencyResponse struct {
	Name string `json:"name"`
}

func toAgencyResponse(m models.Agency) agencyResponse {
	return agencyResponse{
		Name: m.Name,
	}
}
