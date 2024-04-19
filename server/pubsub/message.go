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

type Payloader interface {
	Payload() Payload
}

type Message struct {
	models.Auditable
	Topic   Topic   `json:"topic" db:"topic"`
	Payload Payload `json:"payload" db:"payload"`
	Retries int     `json:"retries" db:"retries"`
}

func NewMessage[P Payloader](topic Topic, payloader P) (Message, error) {
	var message Message
	message.Topic = topic
	message.Payload = payloader.Payload()
	return message, nil
}

func Unmarshal[T any](message Message, out *T) error {
	b, err := json.Marshal(message.Payload)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, out); err != nil {
		return err
	}
	return nil
}

type PayloadProvisionUser struct {
	AgencyID string      `json:"agencyId"`
	Email    string      `json:"email"`
	Role     models.Role `json:"role"`
}

func (p PayloadProvisionUser) Payload() Payload {
	var payload Payload = make(Payload)
	payload["agencyId"] = p.AgencyID
	payload["email"] = p.Email
	payload["role"] = p.Role
	return payload
}

type PayloadSendPage struct {
	PageDeliveryID string `json:"pageDeliveryId"`
}

func (p PayloadSendPage) Payload() Payload {
	var payload Payload = make(Payload)
	payload["pageDeliveryId"] = p.PageDeliveryID
	return payload
}
