package worker

import "log/slog"

type Config struct {
	LogLevel          slog.Level `env:"LOG_LEVEL"`
	Environment       string     `env:"ENVIRONMENT"`
	EndpointTableName string     `env:"ENDPOINT_TABLE_NAME"`
	EventsTopicARN    string     `env:"EVENTS_TOPIC_ARN"`
}
