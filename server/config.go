package main

import (
	"fmt"
	"strings"
)

type environment string

const (
	environmentDev environment = "development"
	enviromentProd environment = "production"
)

type config struct {
	Environment     environment
	Host            string
	Port            string
	DBConn          string
	Auth0Audience   string
	Auth0Domain     string
	SpiceDBEndpoint string
	SpiceDBToken    string
}

func newConfigFromProcessEnv(getenv func(string) string) (config, error) {
	missing := make([]string, 0)
	env := getenv("ENVIRONMENT")
	if env == "" {
		env = string(environmentDev)
	}

	host := getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}

	port := getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbConn := getenv("DB_CONN")
	if dbConn == "" {
		missing = append(missing, "DB_CONN")
	}

	auth0Audience := getenv("AUTH0_AUDIENCE")
	if auth0Audience == "" {
		missing = append(missing, "AUTH0_AUDIENCE")
	}

	auth0Domain := getenv("AUTH0_DOMAIN")
	if auth0Domain == "" {
		missing = append(missing, "AUTH0_DOMAIN")
	}

	spiceDBEndpoint := getenv("SPICEDB_ENDPOINT")
	if spiceDBEndpoint == "" {
		missing = append(missing, "SPICEDB_ENDPOINT")
	}

	spiceDBToken := getenv("SPICEDB_TOKEN")
	if spiceDBToken == "" {
		missing = append(missing, "SPICEDB_TOKEN")
	}

	if len(missing) > 0 {
		return config{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ","))
	}

	return config{
		Environment:     environment(env),
		Host:            host,
		Port:            port,
		DBConn:          dbConn,
		Auth0Audience:   auth0Audience,
		Auth0Domain:     auth0Domain,
		SpiceDBEndpoint: spiceDBEndpoint,
		SpiceDBToken:    spiceDBToken,
	}, nil
}
