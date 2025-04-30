package app

import "log/slog"

type Config struct {
	Auth0Domain     string     `env:"AUTH0_DOMAIN"`
	Auth0Audience   string     `env:"AUTH0_AUDIENCE"`
	UserTableName   string     `env:"USER_TABLE_NAME"`
	AgencyTableName string     `env:"AGENCY_TABLE_NAME"`
	LogLevel        slog.Level `env:"LOG_LEVEL" default:"ERROR"`
}
