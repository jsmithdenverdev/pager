package app

type Config struct {
	Environment     string `env:"ENVIRONMENT"`
	AgencyTableName string `env:"AGENCY_TABLE_NAME"`
}
