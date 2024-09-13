package handlers

import (
	"time"

	"github.com/jsmithdenverdev/pager/models"
)

// deviceResponse is the response model representation of a models.Device.
type deviceResponse struct {
	ID         string    `json:"id"`
	Created    time.Time `json:"created"`
	CreatedBy  string    `json:"createdBy"`
	Modified   time.Time `json:"modified"`
	ModifiedBy string    `json:"modifiedBy"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	Endpoint   string    `json:"endpoint"`
	UserID     string    `json:"userId"`
	Code       string    `json:"code"`
}

// toDeviceResponse converts a models.Device to a deviceResponse.
func toDeviceResponse(m models.Device) deviceResponse {
	return deviceResponse{
		ID:         m.ID,
		Created:    m.Created,
		CreatedBy:  m.CreatedBy,
		Modified:   m.Modified,
		ModifiedBy: m.ModifiedBy,
		Name:       m.Name,
		Status:     string(m.Status),
		Endpoint:   m.Endpoint,
		UserID:     m.UserID,
		Code:       m.Code,
	}
}
