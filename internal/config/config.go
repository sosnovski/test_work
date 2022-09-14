package config

import (
	"fmt"

	"github.com/vrischmann/envconfig"
)

type Config struct {
	PostgresDSN string
	Addr        string
	SeamlessURI string
}

func Init(prefix string) (Config, error) {
	config := Config{}
	if err := envconfig.InitWithPrefix(&config, prefix); err != nil {
		return config, fmt.Errorf("failed to init config: %w", err)
	}

	return config, nil
}
