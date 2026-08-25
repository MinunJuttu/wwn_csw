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
	AdminPassword string
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

	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		return Config{}, fmt.Errorf("ADMIN_PASSWORD is required")
	}

	return Config{
		Environment:   environment,
		ServerAddress: serverAddress,
		SecureCookies: environment == EnvironmentProduction,
		AdminPassword: adminPassword,
	}, nil
}
