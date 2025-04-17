package app

type Config struct {
	PolicyStoreID string `env:"POLICY_STORE_ID"`
	TableName     string `env:"TABLE_NAME"`
}
