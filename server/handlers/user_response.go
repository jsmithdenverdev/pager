package handlers

import (
	"time"

	"github.com/jsmithdenverdev/pager/models"
)

// userResponse is the response model representation of a models.User.
type userResponse struct {
	ID         string    `json:"id"`
	Created    time.Time `json:"created"`
	CreatedBy  string    `json:"createdBy"`
	Modified   time.Time `json:"modified"`
	ModifiedBy string    `json:"modifiedBy"`
	Email      string    `json:"email"`
	IdpID      string    `json:"idpId"`
	Status     string    `json:"status"`
}

// toUserResponse converts a models.User to a userResponse
func toUserResponse(m models.User) userResponse {
	return userResponse{
		ID:         m.ID,
		Created:    m.Created,
		CreatedBy:  m.CreatedBy,
		Modified:   m.Modified,
		ModifiedBy: m.ModifiedBy,
		Email:      m.Email,
		IdpID:      m.IdpID,
		Status:     string(m.Status),
	}
}
