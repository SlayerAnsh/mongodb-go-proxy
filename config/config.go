package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	MongoURI          string
	APISecret         string
	ReadOnlyAPISecret string
	ServerPort        string
	Database          string
}

// Load reads configuration from environment variables and .env file
func Load() *Config {
	// Try to load .env file (ignore error if file doesn't exist)
	if err := godotenv.Load(); err != nil {
		// Only log if .env file was explicitly looked for but not found
		// This allows the app to work with just environment variables
		log.Println("No .env file found, using environment variables only")
	}

	return &Config{
		MongoURI:          GetEnv("MONGO_URI", ""),
		APISecret:         GetEnv("API_SECRET", ""),
		ReadOnlyAPISecret: GetEnv("READONLY_API_SECRET", ""),
		ServerPort:        GetEnv("PORT", "8080"),
		Database:          GetEnv("MONGO_DATABASE", ""),
	}
}

// getEnv retrieves an environment variable or returns a default value
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Validate checks if required configuration is present
func (c *Config) Validate() error {
	if c.MongoURI == "" {
		return &ConfigError{Field: "MONGO_URI", Message: "MongoDB URI is required"}
	}
	if c.APISecret == "" {
		return &ConfigError{Field: "API_SECRET", Message: "API Secret is required"}
	}
	return nil
}

// ConfigError represents a configuration error
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}
