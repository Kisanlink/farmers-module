package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	ServiceName   string
	Environment   string
	Database      DatabaseConfig
	Server        ServerConfig
	AAA           AAAConfig
	Observability ObservabilityConfig
}

// DatabaseConfig holds database configuration matching kisanlink-db
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	MaxConns int
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string
	Host string
}

// AAAConfig holds AAA service configuration
type AAAConfig struct {
	GRPCEndpoint    string
	Token           string
	APIKey          string
	RetryAttempts   int
	RetryBackoff    string
	RequestTimeout  string
	Enabled         bool
	JWTSecret       string
	JWTPublicKey    string
	DefaultPassword string
}

// ObservabilityConfig holds observability configuration
type ObservabilityConfig struct {
	LogLevel                 string
	EnableTracing            bool
	EnableMetrics            bool
	OTELExporterOTLPEndpoint string
}

// Load loads configuration from environment variables
func Load() *Config {
	// Load .env file if it exists (ignore error if file doesn't exist)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config := &Config{
		ServiceName: getEnv("SERVICE_NAME", "farmers-module"),
		Environment: getEnv("ENVIRONMENT", "development"),
		Database: DatabaseConfig{
			Host:     getEnv("DB_POSTGRES_HOST", "localhost"),
			Port:     getEnv("DB_POSTGRES_PORT", "5432"),
			User:     getEnv("DB_POSTGRES_USER", "postgres"),
			Password: getEnv("DB_POSTGRES_PASSWORD", "postgres"),
			Name:     getEnv("DB_POSTGRES_DBNAME", "farmers_module"),
			SSLMode:  getEnv("DB_POSTGRES_SSLMODE", "disable"),
			MaxConns: getEnvAsInt("DB_POSTGRES_MAX_CONNS", 10),
		},
		Server: ServerConfig{
			Port: getEnv("SERVICE_PORT", "8000"),
			Host: getEnv("HOST", "0.0.0.0"),
		},
		AAA: AAAConfig{
			GRPCEndpoint:    getEnv("AAA_GRPC_ADDR", "localhost:50051"),
			Token:           getEnv("AAA_TOKEN", ""),
			APIKey:          getEnv("AAA_API_KEY", ""),
			RetryAttempts:   getEnvAsInt("AAA_RETRY_ATTEMPTS", 3),
			RetryBackoff:    getEnv("AAA_RETRY_BACKOFF", "100ms"),
			RequestTimeout:  getEnv("AAA_REQUEST_TIMEOUT", "5s"),
			Enabled:         getEnvAsBool("AAA_ENABLED", true),
			JWTSecret:       getEnv("JWT_SECRET", "dev-secret-change-in-production"),
			JWTPublicKey:    getEnv("JWT_PUBLIC_KEY", ""),
			DefaultPassword: getEnv("AAA_DEFAULT_PASSWORD", "Welcome@123"),
		},
		Observability: ObservabilityConfig{
			LogLevel:                 getEnv("LOG_LEVEL", "info"),
			EnableTracing:            getEnvAsBool("ENABLE_TRACING", true),
			EnableMetrics:            getEnvAsBool("ENABLE_METRICS", true),
			OTELExporterOTLPEndpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		},
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// Log configuration (without sensitive data)
	log.Printf("Configuration loaded successfully:")
	log.Printf("  Service: %s", config.ServiceName)
	log.Printf("  Environment: %s", config.Environment)
	log.Printf("  Server: %s:%s", config.Server.Host, config.Server.Port)
	log.Printf("  Database: %s:%s/%s", config.Database.Host, config.Database.Port, config.Database.Name)
	log.Printf("  AAA Service: %s", config.AAA.GRPCEndpoint)

	return config
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.Database.Port == "" {
		return fmt.Errorf("DB_PORT is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("DB_USER is required")
	}
	if c.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	if c.Server.Port == "" {
		return fmt.Errorf("SERVICE_PORT is required")
	}
	return nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as integer with a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool gets an environment variable as boolean with a default value
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
