package core_test

import (
	"context"
	"fmt"
	"io"
	"log"
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
