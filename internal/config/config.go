package config

import (
	"fmt"
	"os"
)

const (
	EnvironmentDevelopment = "development"
	EnvironmentProduction  = "production"
)

type Config struct {
	Environment   string
	ServerAddress string
	SecureCookies bool
}

func Load() (Config, error) {
	environment := os.Getenv("APP_ENV")

	if environment == "" {
		environment = EnvironmentDevelopment
	}

	if environment != EnvironmentDevelopment &&
		environment != EnvironmentProduction {
		return Config{}, fmt.Errorf(
			"invalid APP_ENV %q: expected %q or %q",
			environment,
			EnvironmentDevelopment,
			EnvironmentProduction,
		)
	}

	serverAddress := os.Getenv("SERVER_ADDRESS")

	if serverAddress == "" {
		serverAddress = "127.0.0.1:8080"
	}

	return Config{
		Environment:   environment,
		ServerAddress: serverAddress,
		SecureCookies: environment == EnvironmentProduction,
	}, nil
}
