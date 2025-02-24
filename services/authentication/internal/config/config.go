package config

import (
	"fmt"
	"strings"
)

type Config struct {
	TableName     string
	Auth0Domain   string
	Auth0Audience string
}

func LoadFromEnv(getenv func(string) string) (Config, error) {
	missing := make([]string, 0)
	tableName := getenv("TABLE_NAME")
	if tableName == "" {
		missing = append(missing, "TABLE_NAME")
	}

	auth0Domain := getenv("AUTH0_DOMAIN")
	if auth0Domain == "" {
		missing = append(missing, "AUTH0_DOMAIN")
	}

	auth0Audience := getenv("AUTH0_AUDIENCE")
	if auth0Audience == "" {
		missing = append(missing, "AUTH0_AUDIENCE")
	}

	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ","))
	}

	return Config{
		Auth0Domain:   auth0Domain,
		Auth0Audience: auth0Audience,
		TableName:     tableName,
	}, nil
}
