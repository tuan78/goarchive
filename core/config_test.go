package core_test

import (
	"os"
	"testing"

	"goarchive/core"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *core.Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with disk storage",
			config: &core.Config{
				Database: core.DatabaseConfig{
					Host:     "localhost",
					Username: "postgres",
					Port:     5432,
				},
				Storage: core.StorageConfig{
					Type: "disk",
					Path: "./backups",
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with S3 storage",
			config: &core.Config{
				Database: core.DatabaseConfig{
					Host:     "localhost",
					Username: "postgres",
					Port:     5432,
				},
				Storage: core.StorageConfig{
					Type:   "s3",
					Bucket: "my-bucket",
					Region: "us-east-1",
				},
			},
			wantErr: false,
		},
		{
			name: "missing database host",
			config: &core.Config{
				Database: core.DatabaseConfig{
					Username: "postgres",
					Port:     5432,
				},
				Storage: core.StorageConfig{
					Type: "disk",
				},
			},
			wantErr: true,
			errMsg:  "database host is required",
		},
		{
			name: "missing database username",
			config: &core.Config{
				Database: core.DatabaseConfig{
					Host: "localhost",
					Port: 5432,
				},
				Storage: core.StorageConfig{
					Type: "disk",
				},
			},
			wantErr: true,
			errMsg:  "database username is required",
		},
		{
			name: "S3 storage missing bucket",
			config: &core.Config{
				Database: core.DatabaseConfig{
					Host:     "localhost",
					Username: "postgres",
					Port:     5432,
				},
				Storage: core.StorageConfig{
					Type:   "s3",
					Region: "us-east-1",
				},
			},
			wantErr: true,
			errMsg:  "storage bucket is required for S3 storage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("Config.Validate() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
		check   func(*testing.T, *core.Config)
	}{
		{
			name: "default values",
			envVars: map[string]string{
				"DB_USERNAME": "testuser",
			},
			wantErr: false,
			check: func(t *testing.T, cfg *core.Config) {
				if cfg.Database.Type != "postgres" {
					t.Errorf("expected DB type 'postgres', got %v", cfg.Database.Type)
				}
				if cfg.Database.Host != "localhost" {
					t.Errorf("expected host 'localhost', got %v", cfg.Database.Host)
				}
				if cfg.Database.Port != 5432 {
					t.Errorf("expected port 5432, got %v", cfg.Database.Port)
				}
				if cfg.Storage.Type != "disk" {
					t.Errorf("expected storage type 'disk', got %v", cfg.Storage.Type)
				}
			},
		},
		{
			name: "custom environment values",
			envVars: map[string]string{
				"DB_TYPE":            "mysql",
				"DB_HOST":            "db.example.com",
				"DB_PORT":            "3306",
				"DB_USERNAME":        "admin",
				"DB_PASSWORD":        "secret",
				"DB_DATABASE":        "myapp",
				"STORAGE_TYPE":       "s3",
				"STORAGE_BUCKET":     "my-backups",
				"STORAGE_REGION":     "us-west-2",
				"STORAGE_ACCESS_KEY": "AKIAIOSFODNN7EXAMPLE",
				"STORAGE_SECRET_KEY": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			},
			wantErr: false,
			check: func(t *testing.T, cfg *core.Config) {
				if cfg.Database.Type != "mysql" {
					t.Errorf("expected DB type 'mysql', got %v", cfg.Database.Type)
				}
				if cfg.Database.Host != "db.example.com" {
					t.Errorf("expected host 'db.example.com', got %v", cfg.Database.Host)
				}
				if cfg.Database.Port != 3306 {
					t.Errorf("expected port 3306, got %v", cfg.Database.Port)
				}
				if cfg.Database.Username != "admin" {
					t.Errorf("expected username 'admin', got %v", cfg.Database.Username)
				}
				if cfg.Storage.Type != "s3" {
					t.Errorf("expected storage type 's3', got %v", cfg.Storage.Type)
				}
				if cfg.Storage.Bucket != "my-backups" {
					t.Errorf("expected bucket 'my-backups', got %v", cfg.Storage.Bucket)
				}
			},
		},
		{
			name: "invalid port number",
			envVars: map[string]string{
				"DB_USERNAME": "testuser",
				"DB_PORT":     "invalid",
			},
			wantErr: false,
			check: func(t *testing.T, cfg *core.Config) {
				if cfg.Database.Port != 5432 {
					t.Errorf("expected default port 5432, got %v", cfg.Database.Port)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()

			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			cfg, err := core.LoadConfigFromEnv()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfigFromEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, cfg)
			}
		})
	}

	os.Clearenv()
}
