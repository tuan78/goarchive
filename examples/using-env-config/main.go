package main

import (
	"context"
	"log"
	"time"

	"goarchive/core"

	// Import plugins
	_ "goarchive/database/postgres"
	_ "goarchive/storage/s3"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Load configuration from environment variables
	// DB_TYPE, DB_HOST, DB_PORT, DB_USERNAME, DB_PASSWORD, DB_DATABASE
	// STORAGE_TYPE, STORAGE_BUCKET, STORAGE_REGION, etc.
	config, err := core.LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create providers using the registry
	db, err := core.GetDatabase(config.Database.Type, &config.Database)
	if err != nil {
		log.Fatalf("Failed to create database provider: %v", err)
	}
	defer db.Close()

	storage, err := core.GetStorage(ctx, config.Storage.Type, &config.Storage)
	if err != nil {
		log.Fatalf("Failed to create storage provider: %v", err)
	}

	// Create backup service
	service := core.NewBackupService(db, storage)

	// Execute backup
	log.Println("Starting backup process...")
	metadata, err := service.Execute(ctx)
	if err != nil {
		log.Fatalf("Backup failed: %v", err)
	}

	log.Printf("✓ Backup completed successfully")
	log.Printf("  ID: %s", metadata.ID)
	log.Printf("  Size: %.2f MB", float64(metadata.Size)/(1024*1024))
	log.Printf("  Checksum: %s", metadata.Checksum)
}
