package app

type Config struct {
	Environment    string `env:"ENVIRONMENT"`
	UserTableName  string `env:"USER_TABLE_NAME"`
	EventsTopicARN string `env:"EVENTS_TOPIC_ARN"`
}
