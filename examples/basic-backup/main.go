package main

import (
	"context"
	"log"
	"time"

	"goarchive/core"

	// Import only the plugins you need
	_ "goarchive/database/postgres"
	_ "goarchive/storage/s3"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Configure database
	dbConfig := &core.DatabaseConfig{
		Type:     "postgres",
		Host:     "localhost",
		Port:     5432,
		Username: "postgres",
		Password: "secret",
		Database: "myapp",
		SSLMode:  "disable",
	}

	// Configure storage
	storageConfig := &core.StorageConfig{
		Type:      "s3",
		Bucket:    "my-backups",
		Region:    "us-west-2",
		AccessKey: "your-access-key",
		SecretKey: "your-secret-key",
		Prefix:    "backups/",
	}

	// Create providers using the registry
	db, err := core.GetDatabase(dbConfig.Type, dbConfig)
	if err != nil {
		log.Fatalf("Failed to create database provider: %v", err)
	}
	defer db.Close()

	storage, err := core.GetStorage(ctx, storageConfig.Type, storageConfig)
	if err != nil {
		log.Fatalf("Failed to create storage provider: %v", err)
	}

	// Create backup service
	service := core.NewBackupService(db, storage)

	// Execute backup
	log.Println("Creating backup...")
	metadata, err := service.Execute(ctx)
	if err != nil {
		log.Fatalf("Backup failed: %v", err)
	}

	log.Printf("Backup completed: %s (%.2f MB)",
		metadata.ID,
		float64(metadata.Size)/(1024*1024))

	// List available backups
	backups, err := service.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list backups: %v", err)
	}

	log.Printf("Found %d backups", len(backups))
	for _, backup := range backups {
		log.Printf("  - %s (%s)", backup.ID, backup.Timestamp.Format(time.RFC3339))
	}

	// Uncomment to restore from a specific backup
	// err = service.Restore(ctx, "backup-id")
	// if err != nil {
	// 	log.Fatalf("Restore failed: %v", err)
	// }
}
