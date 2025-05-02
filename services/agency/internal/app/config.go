package app

import "log/slog"

type Config struct {
	LogLevel        slog.Level `env:"LOG_LEVEL"`
	Environment     string     `env:"ENVIRONMENT"`
	AgencyTableName string     `env:"AGENCY_TABLE_NAME"`
	EventsTopicARN  string     `env:"EVENTS_TOPIC_ARN"`
}
