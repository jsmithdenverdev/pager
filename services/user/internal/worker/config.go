package worker

import "log/slog"

type Config struct {
	LogLevel                    slog.Level `env:"LOG_LEVEL"`
	Environment                 string     `env:"ENVIRONMENT"`
	UserTableName               string     `env:"USER_TABLE_NAME"`
	EventsTopicARN              string     `env:"EVENTS_TOPIC_ARN"`
	EventRetryCount             int        `env:"EVENT_RETRY_COUNT"`
	Auth0Domain                 string     `env:"AUTH0_DOMAIN"`
	Auth0ManagementClientID     string     `env:"AUTH0_MANAGEMENT_CLIENT_ID"`
	Auth0ManagementClientSecret string     `env:"AUTH0_MANAGEMENT_CLIENT_SECRET"`
	Auth0Connection             string     `env:"AUTH0_CONNECTION"`
}
