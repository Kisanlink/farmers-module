package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	AAA      AAAConfig
}

// DatabaseConfig holds database configuration matching kisanlink-db
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string
	Host string
}

// AAAConfig holds AAA service configuration
type AAAConfig struct {
	GRPCEndpoint string
	Token        string
}

// Load loads configuration from environment variables
func Load() *Config {
	// Load .env file if it exists (ignore error if file doesn't exist)
	_ = godotenv.Load()

	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_POSTGRES_HOST", "localhost"),
			Port:     getEnv("DB_POSTGRES_PORT", "5432"),
			User:     getEnv("DB_POSTGRES_USER", "postgres"),
			Password: getEnv("DB_POSTGRES_PASSWORD", "postgres"),
			Name:     getEnv("DB_POSTGRES_DBNAME", "farmers_module"),
			SSLMode:  getEnv("DB_POSTGRES_SSLMODE", "disable"),
		},
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Host: getEnv("HOST", "0.0.0.0"),
		},
		AAA: AAAConfig{
			GRPCEndpoint: getEnv("AAA_GRPC_ENDPOINT", "localhost:50051"),
			Token:        getEnv("AAA_TOKEN", ""),
		},
	}
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
