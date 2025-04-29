package app

type Config struct {
	Environment       string `env:"ENVIRONMENT"`
	EndpointTableName string `env:"ENDPOINT_TABLE_NAME"`
}
