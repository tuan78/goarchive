package core_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"testing"
	"time"

	"goarchive/core"
)

// ExampleNewBackupService demonstrates how to create and use a backup service
func ExampleNewBackupService() {
	ctx := context.Background()

	// Create mock providers for demonstration
	db := &mockDatabaseProvider{}
	storage := &mockStorageProvider{}

	// Create backup service
	service := core.NewBackupService(db, storage)

	// Execute backup
	metadata, err := service.Execute(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Backup created: %s\n", metadata.ID)
	// Output: Backup created: 20060102-150405
}

// ExampleGetDatabase demonstrates how to get a database provider from the registry
func ExampleGetDatabase() {
	// Register a mock provider for demonstration
	core.RegisterDatabase("mock", func(config *core.DatabaseConfig) (core.DatabaseProvider, error) {
		return &mockDatabaseProvider{}, nil
	})

	// Get the provider from registry
	config := &core.DatabaseConfig{
		Type:     "mock",
		Host:     "localhost",
		Port:     5432,
		Database: "mydb",
	}

	provider, err := core.GetDatabase("mock", config)
	if err != nil {
		log.Fatal(err)
	}
	defer provider.Close()

	fmt.Println("Database provider created successfully")
	// Output: Database provider created successfully
}

// ExampleGetStorage demonstrates how to get a storage provider from the registry
func ExampleGetStorage() {
	ctx := context.Background()

	// Register a mock provider for demonstration
	core.RegisterStorage("mock", func(ctx context.Context, config *core.StorageConfig) (core.StorageProvider, error) {
		return &mockStorageProvider{}, nil
	})

	// Get the provider from registry
	config := &core.StorageConfig{
		Type:   "mock",
		Bucket: "my-backups",
		Region: "us-west-2",
	}

	provider, err := core.GetStorage(ctx, "mock", config)
	if err != nil {
		log.Fatal(err)
	}
	_ = provider

	fmt.Println("Storage provider created successfully")
	// Output: Storage provider created successfully
}

// ExampleListDatabases demonstrates how to list registered database providers
func ExampleListDatabases() {
	// Clear and register some providers for demonstration
	registry := core.NewRegistry()
	registry.RegisterDatabase("postgres", nil)
	registry.RegisterDatabase("mysql", nil)

	databases := registry.ListDatabases()
	fmt.Printf("Found %d database providers\n", len(databases))
	// Output: Found 2 database providers
}

// Mock implementations for examples

type mockDatabaseProvider struct{}

func (m *mockDatabaseProvider) Backup(ctx context.Context) (io.ReadCloser, error) {
	return io.NopCloser(nil), nil
}

func (m *mockDatabaseProvider) Restore(ctx context.Context, reader io.Reader) error {
	return nil
}

func (m *mockDatabaseProvider) GetMetadata() (*core.DatabaseMetadata, error) {
	return &core.DatabaseMetadata{
		Type:    "mock",
		Version: "1.0",
		Name:    "testdb",
	}, nil
}

func (m *mockDatabaseProvider) Close() error {
	return nil
}

type mockStorageProvider struct{}

func (m *mockStorageProvider) Upload(ctx context.Context, reader io.Reader, metadata *core.BackupMetadata) error {
	// Set deterministic ID for example output
	metadata.ID = "20060102-150405"
	return nil
}

func (m *mockStorageProvider) List(ctx context.Context) ([]*core.BackupMetadata, error) {
	return []*core.BackupMetadata{
		{
			ID:        "20060102-150405",
			Timestamp: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
		},
	}, nil
}

func (m *mockStorageProvider) Download(ctx context.Context, backupID string) (io.ReadCloser, error) {
	return io.NopCloser(nil), nil
}

func (m *mockStorageProvider) Delete(ctx context.Context, backupID string) error {
	return nil
}

// Test BackupService.Execute

func TestBackupService_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("successful backup", func(t *testing.T) {
		db := &mockDatabaseProvider{}
		storage := &mockStorageProvider{}
		service := core.NewBackupService(db, storage)

		metadata, err := service.Execute(ctx)
		if err != nil {
			t.Errorf("Execute() error = %v, want nil", err)
		}

		if metadata == nil {
			t.Error("Expected metadata, got nil")
		}

		if metadata.DatabaseName != "testdb" {
			t.Errorf("Expected DatabaseName 'testdb', got '%s'", metadata.DatabaseName)
		}

		if metadata.DatabaseType != "mock" {
			t.Errorf("Expected DatabaseType 'mock', got '%s'", metadata.DatabaseType)
		}
	})

	t.Run("GetMetadata error", func(t *testing.T) {
		db := &mockDatabaseProviderWithError{metadataErr: fmt.Errorf("metadata error")}
		storage := &mockStorageProvider{}
		service := core.NewBackupService(db, storage)

		_, err := service.Execute(ctx)
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})

	t.Run("Backup error", func(t *testing.T) {
		db := &mockDatabaseProviderWithError{backupErr: fmt.Errorf("backup error")}
		storage := &mockStorageProvider{}
		service := core.NewBackupService(db, storage)

		_, err := service.Execute(ctx)
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})

	t.Run("Upload error", func(t *testing.T) {
		db := &mockDatabaseProvider{}
		storage := &mockStorageProviderWithError{uploadErr: fmt.Errorf("upload error")}
		service := core.NewBackupService(db, storage)

		_, err := service.Execute(ctx)
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}

func TestBackupService_Restore(t *testing.T) {
	ctx := context.Background()

	t.Run("successful restore", func(t *testing.T) {
		db := &mockDatabaseProvider{}
		storage := &mockStorageProvider{}
		service := core.NewBackupService(db, storage)

		err := service.Restore(ctx, "backup-123")
		if err != nil {
			t.Errorf("Restore() error = %v, want nil", err)
		}
	})

	t.Run("Download error", func(t *testing.T) {
		db := &mockDatabaseProvider{}
		storage := &mockStorageProviderWithError{downloadErr: fmt.Errorf("download error")}
		service := core.NewBackupService(db, storage)

		err := service.Restore(ctx, "backup-123")
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})

	t.Run("Restore error", func(t *testing.T) {
		db := &mockDatabaseProviderWithError{restoreErr: fmt.Errorf("restore error")}
		storage := &mockStorageProvider{}
		service := core.NewBackupService(db, storage)

		err := service.Restore(ctx, "backup-123")
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}

func TestBackupService_List(t *testing.T) {
	ctx := context.Background()

	t.Run("successful list", func(t *testing.T) {
		db := &mockDatabaseProvider{}
		storage := &mockStorageProvider{}
		service := core.NewBackupService(db, storage)

		backups, err := service.List(ctx)
		if err != nil {
			t.Errorf("List() error = %v, want nil", err)
		}

		if len(backups) != 1 {
			t.Errorf("Expected 1 backup, got %d", len(backups))
		}
	})

	t.Run("List error", func(t *testing.T) {
		db := &mockDatabaseProvider{}
		storage := &mockStorageProviderWithError{listErr: fmt.Errorf("list error")}
		service := core.NewBackupService(db, storage)

		_, err := service.List(ctx)
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}

func TestBackupService_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("successful delete", func(t *testing.T) {
		db := &mockDatabaseProvider{}
		storage := &mockStorageProvider{}
		service := core.NewBackupService(db, storage)

		err := service.Delete(ctx, "backup-123")
		if err != nil {
			t.Errorf("Delete() error = %v, want nil", err)
		}
	})

	t.Run("Delete error", func(t *testing.T) {
		db := &mockDatabaseProvider{}
		storage := &mockStorageProviderWithError{deleteErr: fmt.Errorf("delete error")}
		service := core.NewBackupService(db, storage)

		err := service.Delete(ctx, "backup-123")
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}

// Mock providers with error injection

type mockDatabaseProviderWithError struct {
	metadataErr error
	backupErr   error
	restoreErr  error
}

func (m *mockDatabaseProviderWithError) Backup(ctx context.Context) (io.ReadCloser, error) {
	if m.backupErr != nil {
		return nil, m.backupErr
	}
	return io.NopCloser(nil), nil
}

func (m *mockDatabaseProviderWithError) Restore(ctx context.Context, reader io.Reader) error {
	return m.restoreErr
}

func (m *mockDatabaseProviderWithError) GetMetadata() (*core.DatabaseMetadata, error) {
	if m.metadataErr != nil {
		return nil, m.metadataErr
	}
	return &core.DatabaseMetadata{
		Type:    "mock",
		Version: "1.0",
		Name:    "testdb",
	}, nil
}

func (m *mockDatabaseProviderWithError) Close() error {
	return nil
}

type mockStorageProviderWithError struct {
	uploadErr   error
	listErr     error
	downloadErr error
	deleteErr   error
}

func (m *mockStorageProviderWithError) Upload(ctx context.Context, reader io.Reader, metadata *core.BackupMetadata) error {
	if m.uploadErr != nil {
		return m.uploadErr
	}
	metadata.ID = "20060102-150405"
	return nil
}

func (m *mockStorageProviderWithError) List(ctx context.Context) ([]*core.BackupMetadata, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return []*core.BackupMetadata{
		{
			ID:        "20060102-150405",
			Timestamp: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
		},
	}, nil
}

func (m *mockStorageProviderWithError) Download(ctx context.Context, backupID string) (io.ReadCloser, error) {
	if m.downloadErr != nil {
		return nil, m.downloadErr
	}
	return io.NopCloser(nil), nil
}

func (m *mockStorageProviderWithError) Delete(ctx context.Context, backupID string) error {
	return m.deleteErr
}
