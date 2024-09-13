package handlers

import "github.com/jsmithdenverdev/pager/models"

// agencyResponse is the response model representation of a models.Agency.
type agencyResponse struct {
	Name string `json:"name"`
}

// toAgencyResponse converts a models.Agency to an agencyResponse.
func toAgencyResponse(m models.Agency) agencyResponse {
	return agencyResponse{
		Name: m.Name,
	}
}
