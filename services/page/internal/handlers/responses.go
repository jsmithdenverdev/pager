package handlers

import (
	"time"

	"github.com/jsmithdenverdev/pager/services/page/internal/models"
)

type pageResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Created     time.Time `json:"created"`
	CreatedBy   string    `json:"createdBy"`
	Modified    time.Time `json:"modified"`
	ModifiedBy  string    `json:"modifiedBy"`
	Location    struct {
		Type models.LocationType `json:"type"`
		Data string              `json:"data"`
	} `json:"location"`
}
