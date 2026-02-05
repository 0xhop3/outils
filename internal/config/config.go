package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port                    string
	Environment             string
	DatabaseURL             string
	FirebaseCredentialsFile string
}

func Load() (*Config, error) {
	config := &Config{
		Port:                    getEnv("PORT", "8080"),
		Environment:             getEnv("ENVIRONMENT", "development"),
		FirebaseCredentialsFile: getEnv("FIREBASE_CREDENTIALS_FILE", "firebase-credentials.json"),
		DatabaseURL:             os.Getenv("DATABASE_URL"),
	}

	if config.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

func (c *Config) IsDev() bool {
	return c.Environment == "development"
}

func (c *Config) IsProd() bool {
	return c.Environment == "production"
}
