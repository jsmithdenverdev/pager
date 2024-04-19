package pubsub

import (
	"encoding/json"

	"github.com/jsmithdenverdev/pager/models"
)

type Message struct {
	models.Auditable
	Topic   Topic  `json:"topic" db:"topic"`
	payload []byte `json:"payload" db:"payload"`
	Retries int    `json:"retries" db:"retries"`
}

func NewMessage[P any](topic Topic, payload P) (Message, error) {
	var message Message
	message.Topic = topic
	payloadB, err := json.Marshal(payload)
	if err != nil {
		return message, err
	}
	message.payload = payloadB
	return message, nil
}

func Unmarshal[T any](message Message, out *T) error {
	if err := json.Unmarshal(message.payload, out); err != nil {
		return err
	}
	return nil
}

type PayloadProvisionUser struct {
	AgencyID string      `json:"agencyId"`
	Email    string      `json:"email"`
	Role     models.Role `json:"role"`
}

type PayloadSendPage struct {
	PageDeliveryID string `json:"pageDeliveryId"`
}
