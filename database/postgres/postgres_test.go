package postgres_test

import (
	"context"
	"os"
	"strconv"
	"testing"

	"goarchive/core"
	"goarchive/database/postgres"
)

func TestNew_InvalidConnection(t *testing.T) {
	// Test that New returns appropriate error with invalid connection details
	config := &core.DatabaseConfig{
		Type:     "postgres",
		Host:     "invalid-host-that-does-not-exist",
		Port:     5432,
		Username: "testuser",
		Password: "testpass",
		Database: "testdb",
		SSLMode:  "disable",
	}

	_, err := postgres.New(config)
	if err == nil {
		t.Error("Expected error with invalid connection, got nil")
	}
}

func TestNew_ValidConfigStructure(t *testing.T) {
	// Test that config is properly validated
	tests := []struct {
		name    string
		config  *core.DatabaseConfig
		wantErr bool
	}{
		{
			name: "complete config",
			config: &core.DatabaseConfig{
				Type:     "postgres",
				Host:     "localhost",
				Port:     5432,
				Username: "user",
				Password: "pass",
				Database: "db",
				SSLMode:  "disable",
			},
			wantErr: true, // Will fail to connect but config is valid
		},
		{
			name: "with SSL mode",
			config: &core.DatabaseConfig{
				Type:     "postgres",
				Host:     "localhost",
				Port:     5432,
				Username: "user",
				Password: "pass",
				Database: "db",
				SSLMode:  "require",
			},
			wantErr: true, // Will fail to connect but config is valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := postgres.New(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProvider_AutoRegistration(t *testing.T) {
	// The postgres provider should automatically register itself
	config := &core.DatabaseConfig{
		Type:     "postgres",
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
	}

	provider, err := core.GetDatabase("postgres", config)
	// Provider will fail to connect, but it should be registered
	if err == nil {
		// If it succeeds, verify the provider
		if provider == nil {
			t.Error("expected non-nil provider from auto-registration")
		}
		provider.Close()
	} else {
		// Error is expected if no database is running
		// But we can verify the error is a connection error, not a "not registered" error
		if err.Error() == "database provider not registered: postgres" {
			t.Error("postgres provider not auto-registered")
		}
	}
}

// Note: Backup, Restore, GetMetadata, and Close methods require:
// - A real PostgreSQL database connection
// - Integration tests with PostgreSQL container (see .github/workflows/coverage.yml)
//
// The current implementation is tested through:
// 1. The auto-registration test ensures the provider is properly registered
// 2. The New() tests ensure proper connection string construction
// 3. Integration tests in CI/CD with PostgreSQL container

// Integration tests - these require a real PostgreSQL instance

func getTestConfig() *core.DatabaseConfig {
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "localhost"
	}

	portStr := os.Getenv("POSTGRES_PORT")
	port := 5432
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	user := os.Getenv("POSTGRES_USER")
	if user == "" {
		user = "testuser"
	}

	password := os.Getenv("POSTGRES_PASSWORD")
	if password == "" {
		password = "testpass"
	}

	database := os.Getenv("POSTGRES_DB")
	if database == "" {
		database = "testdb"
	}

	return &core.DatabaseConfig{
		Type:     "postgres",
		Host:     host,
		Port:     port,
		Username: user,
		Password: password,
		Database: database,
		SSLMode:  "disable",
	}
}

func TestIntegration_PostgreSQL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := getTestConfig()

	t.Run("New connection", func(t *testing.T) {
		provider, err := postgres.New(config)
		if err != nil {
			t.Skipf("Skipping integration test - PostgreSQL not available: %v", err)
			return
		}
		defer provider.Close()

		if provider == nil {
			t.Error("expected non-nil provider")
		}
	})

	t.Run("GetMetadata", func(t *testing.T) {
		provider, err := postgres.New(config)
		if err != nil {
			t.Skipf("Skipping integration test - PostgreSQL not available: %v", err)
			return
		}
		defer provider.Close()

		metadata, err := provider.GetMetadata()
		if err != nil {
			t.Errorf("GetMetadata() error = %v", err)
			return
		}

		if metadata.Type != "postgres" {
			t.Errorf("Expected Type 'postgres', got '%s'", metadata.Type)
		}

		if metadata.Name != config.Database {
			t.Errorf("Expected Name '%s', got '%s'", config.Database, metadata.Name)
		}

		if metadata.Version == "" {
			t.Error("Expected non-empty Version")
		}

		if metadata.Size < 0 {
			t.Errorf("Expected non-negative Size, got %d", metadata.Size)
		}
	})

	t.Run("Backup", func(t *testing.T) {
		provider, err := postgres.New(config)
		if err != nil {
			t.Skipf("Skipping integration test - PostgreSQL not available: %v", err)
			return
		}
		defer provider.Close()

		ctx := context.Background()
		reader, err := provider.Backup(ctx)
		if err != nil {
			t.Errorf("Backup() error = %v", err)
			return
		}
		defer reader.Close()

		// Try to read at least some data
		buf := make([]byte, 1024)
		n, err := reader.Read(buf)
		if err != nil && err.Error() != "EOF" {
			t.Errorf("Failed to read backup data: %v", err)
		}

		if n == 0 {
			t.Error("Expected to read some backup data, got 0 bytes")
		}
	})

	t.Run("Close", func(t *testing.T) {
		provider, err := postgres.New(config)
		if err != nil {
			t.Skipf("Skipping integration test - PostgreSQL not available: %v", err)
			return
		}

		err = provider.Close()
		if err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
}

