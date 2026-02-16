package core_test

import (
	"context"
	"errors"
	"testing"

	"goarchive/core"
)

func TestRegistry_RegisterDatabase(t *testing.T) {
	registry := core.NewRegistry()

	factory := func(config *core.DatabaseConfig) (core.DatabaseProvider, error) {
		return &mockDatabaseProvider{}, nil
	}

	registry.RegisterDatabase("test", factory)

	databases := registry.ListDatabases()
	if len(databases) != 1 {
		t.Errorf("expected 1 database, got %d", len(databases))
	}
	if databases[0] != "test" {
		t.Errorf("expected database 'test', got %v", databases[0])
	}
}

func TestRegistry_RegisterStorage(t *testing.T) {
	registry := core.NewRegistry()

	factory := func(ctx context.Context, config *core.StorageConfig) (core.StorageProvider, error) {
		return &mockStorageProvider{}, nil
	}

	registry.RegisterStorage("test", factory)

	storages := registry.ListStorages()
	if len(storages) != 1 {
		t.Errorf("expected 1 storage, got %d", len(storages))
	}
	if storages[0] != "test" {
		t.Errorf("expected storage 'test', got %v", storages[0])
	}
}

func TestRegistry_GetDatabase(t *testing.T) {
	tests := []struct {
		name       string
		register   bool
		returnErr  bool
		wantErr    bool
		errContains string
	}{
		{
			name:     "successful get",
			register: true,
			wantErr:  false,
		},
		{
			name:        "provider not registered",
			register:    false,
			wantErr:     true,
			errContains: "not registered",
		},
		{
			name:        "factory returns error",
			register:    true,
			returnErr:   true,
			wantErr:     true,
			errContains: "factory error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := core.NewRegistry()

			if tt.register {
				factory := func(config *core.DatabaseConfig) (core.DatabaseProvider, error) {
					if tt.returnErr {
						return nil, errors.New("factory error")
					}
					return &mockDatabaseProvider{}, nil
				}
				registry.RegisterDatabase("test", factory)
			}

			config := &core.DatabaseConfig{
				Host:     "localhost",
				Username: "test",
			}

			provider, err := registry.GetDatabase("test", config)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDatabase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || !contains(err.Error(), tt.errContains) {
					t.Errorf("GetDatabase() error = %v, should contain %v", err, tt.errContains)
				}
			}

			if !tt.wantErr && provider == nil {
				t.Error("expected non-nil provider")
			}
		})
	}
}

func TestRegistry_GetStorage(t *testing.T) {
	tests := []struct {
		name        string
		register    bool
		returnErr   bool
		wantErr     bool
		errContains string
	}{
		{
			name:     "successful get",
			register: true,
			wantErr:  false,
		},
		{
			name:        "provider not registered",
			register:    false,
			wantErr:     true,
			errContains: "not registered",
		},
		{
			name:        "factory returns error",
			register:    true,
			returnErr:   true,
			wantErr:     true,
			errContains: "factory error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := core.NewRegistry()
			ctx := context.Background()

			if tt.register {
				factory := func(ctx context.Context, config *core.StorageConfig) (core.StorageProvider, error) {
					if tt.returnErr {
						return nil, errors.New("factory error")
					}
					return &mockStorageProvider{}, nil
				}
				registry.RegisterStorage("test", factory)
			}

			config := &core.StorageConfig{
				Type: "test",
				Path: "./test",
			}

			provider, err := registry.GetStorage(ctx, "test", config)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStorage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || !contains(err.Error(), tt.errContains) {
					t.Errorf("GetStorage() error = %v, should contain %v", err, tt.errContains)
				}
			}

			if !tt.wantErr && provider == nil {
				t.Error("expected non-nil provider")
			}
		})
	}
}

func TestRegistry_ListDatabases(t *testing.T) {
	registry := core.NewRegistry()

	// Initially empty
	if len(registry.ListDatabases()) != 0 {
		t.Error("expected empty database list")
	}

	// Add multiple databases
	factory := func(config *core.DatabaseConfig) (core.DatabaseProvider, error) {
		return &mockDatabaseProvider{}, nil
	}

	registry.RegisterDatabase("postgres", factory)
	registry.RegisterDatabase("mysql", factory)
	registry.RegisterDatabase("mongo", factory)

	databases := registry.ListDatabases()
	if len(databases) != 3 {
		t.Errorf("expected 3 databases, got %d", len(databases))
	}

	// Check all databases are present
	dbMap := make(map[string]bool)
	for _, db := range databases {
		dbMap[db] = true
	}

	expected := []string{"postgres", "mysql", "mongo"}
	for _, exp := range expected {
		if !dbMap[exp] {
			t.Errorf("expected database %v to be in list", exp)
		}
	}
}

func TestRegistry_ListStorages(t *testing.T) {
	registry := core.NewRegistry()

	// Initially empty
	if len(registry.ListStorages()) != 0 {
		t.Error("expected empty storage list")
	}

	// Add multiple storages
	factory := func(ctx context.Context, config *core.StorageConfig) (core.StorageProvider, error) {
		return &mockStorageProvider{}, nil
	}

	registry.RegisterStorage("disk", factory)
	registry.RegisterStorage("s3", factory)
	registry.RegisterStorage("azure", factory)

	storages := registry.ListStorages()
	if len(storages) != 3 {
		t.Errorf("expected 3 storages, got %d", len(storages))
	}

	// Check all storages are present
	storageMap := make(map[string]bool)
	for _, s := range storages {
		storageMap[s] = true
	}

	expected := []string{"disk", "s3", "azure"}
	for _, exp := range expected {
		if !storageMap[exp] {
			t.Errorf("expected storage %v to be in list", exp)
		}
	}
}

func TestDefaultRegistry(t *testing.T) {
	// Test that default registry functions work
	factory := func(config *core.DatabaseConfig) (core.DatabaseProvider, error) {
		return &mockDatabaseProvider{}, nil
	}

	core.RegisterDatabase("test-global", factory)

	databases := core.ListDatabases()
	found := false
	for _, db := range databases {
		if db == "test-global" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected 'test-global' in default registry")
	}

	config := &core.DatabaseConfig{
		Host:     "localhost",
		Username: "test",
	}

	provider, err := core.GetDatabase("test-global", config)
	if err != nil {
		t.Errorf("GetDatabase() from default registry failed: %v", err)
	}
	if provider == nil {
		t.Error("expected non-nil provider from default registry")
	}
}

func TestRegistry_Concurrent(t *testing.T) {
	registry := core.NewRegistry()

	// Register factories
	dbFactory := func(config *core.DatabaseConfig) (core.DatabaseProvider, error) {
		return &mockDatabaseProvider{}, nil
	}

	storageFactory := func(ctx context.Context, config *core.StorageConfig) (core.StorageProvider, error) {
		return &mockStorageProvider{}, nil
	}

	registry.RegisterDatabase("test", dbFactory)
	registry.RegisterStorage("test", storageFactory)

	// Concurrent reads
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			config := &core.DatabaseConfig{Host: "localhost", Username: "test"}
			_, _ = registry.GetDatabase("test", config)
			_ = registry.ListDatabases()
			done <- true
		}()

		go func() {
			ctx := context.Background()
			config := &core.StorageConfig{Type: "test"}
			_, _ = registry.GetStorage(ctx, "test", config)
			_ = registry.ListStorages()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		(s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || findInString(s, substr)))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

