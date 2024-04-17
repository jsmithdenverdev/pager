package pubsub

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/jsmithdenverdev/pager/models"
)

type Payload map[string]interface{}

func (p Payload) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *Payload) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &p)
}

type Message struct {
	models.Auditable
	Topic   Topic   `json:"topic" db:"topic"`
	Payload Payload `json:"payload" db:"payload"`
	Retries int     `json:"retries" db:"retries"`
}

type MessageProvisionUser struct {
	Email string `json:"email"`
}
