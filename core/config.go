package core

import (
	"fmt"
	"os"
)

// Config holds the application configuration
type Config struct {
	Database DatabaseConfig
	Storage  StorageConfig
}

// DatabaseConfig contains database connection settings
type DatabaseConfig struct {
	Type     string
	Host     string
	Port     int
	Username string
	Password string
	Database string
	SSLMode  string
}

// StorageConfig contains storage settings
type StorageConfig struct {
	Type      string
	Bucket    string // For S3-compatible storage
	Region    string // For S3-compatible storage
	Endpoint  string // For S3-compatible storage (e.g., LocalStack)
	AccessKey string // For S3-compatible storage
	SecretKey string // For S3-compatible storage
	Prefix    string // For S3-compatible storage
	Path      string // For disk storage
}

// LoadConfigFromEnv loads configuration from environment variables
// Uses generic DB_* and STORAGE_* prefixes for provider-agnostic configuration
func LoadConfigFromEnv() (*Config, error) {
	config := &Config{
		Database: DatabaseConfig{
			Type:     getEnv("DB_TYPE", "postgres"),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			Username: getEnv("DB_USERNAME", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Database: getEnv("DB_DATABASE", "postgres"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Storage: StorageConfig{
			Type:      getEnv("STORAGE_TYPE", "disk"),
			Bucket:    getEnv("STORAGE_BUCKET", ""),
			Endpoint:  getEnv("STORAGE_ENDPOINT", ""),
			Region:    getEnv("STORAGE_REGION", "us-east-1"),
			AccessKey: getEnv("STORAGE_ACCESS_KEY", ""),
			SecretKey: getEnv("STORAGE_SECRET_KEY", ""),
			Prefix:    getEnv("STORAGE_PREFIX", "backups/"),
			Path:      getEnv("STORAGE_PATH", "./backups"),
		},
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if c.Database.Username == "" {
		return fmt.Errorf("database username is required")
	}

	// Storage validation depends on type
	switch c.Storage.Type {
	case "s3":
		if c.Storage.Bucket == "" {
			return fmt.Errorf("storage bucket is required for S3 storage")
		}
	case "disk":
		// Path is optional, will default to ./backups
		// No validation needed
	default:
		// For unknown storage types, let the provider handle validation
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	var value int
	_, err := fmt.Sscanf(valueStr, "%d", &value)
	if err != nil {
		return defaultValue
	}

	return value
}
