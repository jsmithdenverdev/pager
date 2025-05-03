package app

import "log/slog"

type Config struct {
	LogLevel        slog.Level `env:"LOG_LEVEL"`
	Environment     string     `env:"ENVIRONMENT"`
	UserTableName   string     `env:"USER_TABLE_NAME"`
	EventsTopicARN  string     `env:"EVENTS_TOPIC_ARN"`
	EventRetryCount int        `env:"EVENT_RETRY_COUNT"`
}
