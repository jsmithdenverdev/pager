package app

type Config struct {
	TableName     string `env:"TABLE_NAME"`
	Auth0Domain   string `env:"AUTH0_DOMAIN"`
	Auth0Audience string `env:"AUTH0_AUDIENCE"`
}
