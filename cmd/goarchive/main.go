package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"goarchive/core"

	// Import plugins to trigger auto-registration via init()
	_ "goarchive/database/postgres"
	_ "goarchive/storage/s3"
)

func main() {
	log.Println("Starting goarchive...")

	// Load configuration from environment
	config, err := core.LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Initialize database provider using registry
	dbProvider, err := core.GetDatabase(config.Database.Type, &config.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database provider: %v", err)
	}
	defer dbProvider.Close()

	// Initialize storage provider using registry
	storageProvider, err := core.GetStorage(ctx, config.Storage.Type, &config.Storage)
	if err != nil {
		log.Fatalf("Failed to initialize storage provider: %v", err)
	}

	// Create backup service
	backupService := core.NewBackupService(dbProvider, storageProvider)

	// Execute backup
	log.Println("Starting backup process...")
	metadata, err := backupService.Execute(ctx)
	if err != nil {
		log.Fatalf("Backup failed: %v", err)
	}

	// Print backup details
	fmt.Println("\n=== Backup Completed Successfully ===")
	fmt.Printf("Backup ID:       %s\n", metadata.ID)
	fmt.Printf("Database:        %s (%s)\n", metadata.DatabaseName, metadata.DatabaseType)
	fmt.Printf("Timestamp:       %s\n", metadata.Timestamp.Format(time.RFC3339))
	fmt.Printf("Size:            %d bytes (%.2f MB)\n", metadata.Size, float64(metadata.Size)/(1024*1024))
	fmt.Printf("Checksum (MD5):  %s\n", metadata.Checksum)
	fmt.Println("====================================")

	log.Println("Backup completed successfully")
}
