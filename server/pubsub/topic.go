package pubsub

type Topic string

const (
	TopicSendPage      Topic = "SEND_PAGE"
	TopicProvisionUser Topic = "PROVISION_USER"
)

type topicConfig struct {
	Topic          string `json:"topic" db:"topic"`
	RetriesEnabled bool   `json:"retriesEnabled" db:"retries_enabled"`
	Retries        int    `json:"retries" db:"retries"`
}
