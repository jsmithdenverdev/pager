package config

import (
	"fmt"
	"strings"
)

type Config struct {
	PolicyStoreID string
}

func LoadFromEnv(getenv func(string) string) (Config, error) {
	missing := make([]string, 0)
	PolicyStoreID := getenv("POLICY_STORE_ID ")
	if PolicyStoreID == "" {
		missing = append(missing, "POLICY_STORE_ID ")
	}

	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ","))
	}

	return Config{
		PolicyStoreID: PolicyStoreID,
	}, nil
}
