package app

type Config struct {
	Environment     string `env:"ENVIRONMENT"`
	AgencyTableName string `env:"AGENCY_TABLE_NAME"`
	EventsTopicARN  string `env:"EVENTS_TOPIC_ARN"`
}
