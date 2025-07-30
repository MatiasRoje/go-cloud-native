package config

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	// Database configuration
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
}

func LoadConfig() (*Config, error) {
	return &Config{
		DBHost:     getEnvWithDefault("DB_HOST", "localhost"),
		DBPort:     getEnvWithDefault("DB_PORT", "5432"),
		DBUser:     getEnvWithDefault("DB_USER", "gouser"),
		DBPassword: getEnvWithDefault("DB_PASSWORD", "go123"),
		DBName:     getEnvWithDefault("DB_NAME", "go_cloud_native"),
	}, nil
}

func getEnvWithDefault(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
