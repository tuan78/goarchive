package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"goarchive/core"

	// Import plugins to trigger auto-registration via init()
	_ "goarchive/database/postgres"
	_ "goarchive/storage/disk"
	_ "goarchive/storage/s3"
)

const version = "1.0.0"

func main() {
	// Define subcommands
	backupCmd := flag.NewFlagSet("backup", flag.ExitOnError)
	listCmd := flag.NewFlagSet("list", flag.ExitOnError)

	// Database flags (shared)
	var dbHost, dbUser, dbPass, dbName, dbType, dbSSLMode string
	var dbPort int

	// Storage flags (shared)
	var storageType, storageBucket, storageRegion, storageAccessKey, storageSecretKey, storagePrefix, storagePath string

	// Define flags for backup command
	setupDatabaseFlags(backupCmd, &dbHost, &dbUser, &dbPass, &dbName, &dbType, &dbSSLMode, &dbPort)
	setupStorageFlags(backupCmd, &storageType, &storageBucket, &storageRegion, &storageAccessKey, &storageSecretKey, &storagePrefix, &storagePath)

	// Define flags for list command
	setupStorageFlags(listCmd, &storageType, &storageBucket, &storageRegion, &storageAccessKey, &storageSecretKey, &storagePrefix, &storagePath)

	// Check for subcommand
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "backup":
		backupCmd.Parse(os.Args[2:])
		executeBackup(dbHost, dbUser, dbPass, dbName, dbType, dbSSLMode, dbPort,
			storageType, storageBucket, storageRegion, storageAccessKey, storageSecretKey, storagePrefix, storagePath)

	case "list":
		listCmd.Parse(os.Args[2:])
		executeList(storageType, storageBucket, storageRegion, storageAccessKey, storageSecretKey, storagePrefix, storagePath)

	case "providers":
		printProviders()

	case "version", "-v", "--version":
		fmt.Printf("goarchive version %s\n", version)

	case "help", "-h", "--help":
		printUsage()

	default:
		fmt.Printf("Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("goarchive - Database backup and restore tool")
	fmt.Printf("\nVersion: %s\n", version)
	fmt.Println("\nUsage:")
	fmt.Println("  goarchive <command> [flags]")
	fmt.Println("\nCommands:")
	fmt.Println("  backup      Create a database backup")
	fmt.Println("  list        List available backups")
	fmt.Println("  providers   Show available database and storage providers")
	fmt.Println("  version     Show version information")
	fmt.Println("  help        Show this help message")
	fmt.Println("\nExamples:")
	fmt.Println("  # Show available providers")
	fmt.Println("  goarchive providers")
	fmt.Println("\n  # Backup to local disk (default)")
	fmt.Println("  goarchive backup --db-host localhost --db-name mydb")
	fmt.Println("\n  # Backup to disk with custom path")
	fmt.Println("  goarchive backup --db-host localhost --db-name mydb --storage-path /var/backups")
	fmt.Println("\n  # Backup to S3")
	fmt.Println("  goarchive backup --db-host localhost --db-name mydb --storage-type s3 --storage-bucket my-backups")
	fmt.Println("\n  # Backup using environment variables")
	fmt.Println("  export DB_HOST=localhost DB_NAME=mydb STORAGE_BUCKET=my-backups")
	fmt.Println("  goarchive backup")
	fmt.Println("\n  # List backups")
	fmt.Println("  goarchive list --storage-bucket my-backups --storage-region us-east-1")
	fmt.Println("\nFlags inherit from environment variables if not specified.")
	fmt.Println("Run 'goarchive <command> -h' for command-specific flags.")
}

func setupDatabaseFlags(fs *flag.FlagSet, host, user, pass, name, dbType, sslMode *string, port *int) {
	availableDBs := core.ListDatabases()
	dbTypeHelp := fmt.Sprintf("Database type (available: %v)", availableDBs)

	fs.StringVar(host, "db-host", getEnv("DB_HOST", "localhost"), "Database host")
	fs.IntVar(port, "db-port", getEnvAsInt("DB_PORT", 5432), "Database port")
	fs.StringVar(user, "db-user", getEnv("DB_USERNAME", "postgres"), "Database username")
	fs.StringVar(pass, "db-password", getEnv("DB_PASSWORD", ""), "Database password")
	fs.StringVar(name, "db-name", getEnv("DB_DATABASE", "postgres"), "Database name")
	fs.StringVar(dbType, "db-type", getEnv("DB_TYPE", "postgres"), dbTypeHelp)
	fs.StringVar(sslMode, "db-sslmode", getEnv("DB_SSLMODE", "disable"), "SSL mode (disable, require, verify-full)")
}

func setupStorageFlags(fs *flag.FlagSet, storageType, bucket, region, accessKey, secretKey, prefix, path *string) {
	availableStorages := core.ListStorages()
	storageTypeHelp := fmt.Sprintf("Storage type (available: %v)", availableStorages)

	fs.StringVar(storageType, "storage-type", getEnv("STORAGE_TYPE", "disk"), storageTypeHelp)
	fs.StringVar(path, "storage-path", getEnv("STORAGE_PATH", "./backups"), "Storage path for disk storage")
	fs.StringVar(bucket, "storage-bucket", getEnv("STORAGE_BUCKET", ""), "Storage bucket name (for S3)")
	fs.StringVar(region, "storage-region", getEnv("STORAGE_REGION", "us-east-1"), "Storage region (for S3)")
	fs.StringVar(accessKey, "storage-access-key", getEnv("STORAGE_ACCESS_KEY", ""), "Storage access key (for S3, optional with IAM)")
	fs.StringVar(secretKey, "storage-secret-key", getEnv("STORAGE_SECRET_KEY", ""), "Storage secret key (for S3, optional with IAM)")
	fs.StringVar(prefix, "storage-prefix", getEnv("STORAGE_PREFIX", "backups/"), "Storage prefix path (for S3)")
}

func executeBackup(dbHost, dbUser, dbPass, dbName, dbType, dbSSLMode string, dbPort int,
	storageType, storageBucket, storageRegion, storageAccessKey, storageSecretKey, storagePrefix, storagePath string) {

	log.Println("Starting goarchive backup...")

	// Build configuration
	config := &core.Config{
		Database: core.DatabaseConfig{
			Type:     dbType,
			Host:     dbHost,
			Port:     dbPort,
			Username: dbUser,
			Password: dbPass,
			Database: dbName,
			SSLMode:  dbSSLMode,
		},
		Storage: core.StorageConfig{
			Type:      storageType,
			Bucket:    storageBucket,
			Region:    storageRegion,
			AccessKey: storageAccessKey,
			SecretKey: storageSecretKey,
			Prefix:    storagePrefix,
			Path:      storagePath,
		},
	}

	if err := config.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
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

func executeList(storageType, storageBucket, storageRegion, storageAccessKey, storageSecretKey, storagePrefix, storagePath string) {
	log.Println("Listing backups...")

	config := &core.StorageConfig{
		Type:      storageType,
		Bucket:    storageBucket,
		Region:    storageRegion,
		AccessKey: storageAccessKey,
		SecretKey: storageSecretKey,
		Prefix:    storagePrefix,
		Path:      storagePath,
	}

	ctx := context.Background()

	// Initialize storage provider
	storageProvider, err := core.GetStorage(ctx, config.Type, config)
	if err != nil {
		log.Fatalf("Failed to initialize storage provider: %v", err)
	}

	// List backups
	backups, err := storageProvider.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list backups: %v", err)
	}

	if len(backups) == 0 {
		fmt.Println("No backups found.")
		return
	}

	fmt.Printf("\nFound %d backup(s):\n\n", len(backups))
	for i, backup := range backups {
		fmt.Printf("%d. %s\n", i+1, backup.ID)
		fmt.Printf("   Timestamp: %s\n", backup.Timestamp.Format(time.RFC3339))
		fmt.Printf("   Size:      %.2f MB\n", float64(backup.Size)/(1024*1024))
		if backup.Checksum != "" {
			fmt.Printf("   Checksum:  %s\n", backup.Checksum)
		}
		fmt.Println()
	}
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

func printProviders() {
	fmt.Println("Available Providers")
	fmt.Println("==================")

	fmt.Println("\nDatabase Providers:")
	dbProviders := core.ListDatabases()
	if len(dbProviders) == 0 {
		fmt.Println("  (none registered)")
	} else {
		for _, name := range dbProviders {
			fmt.Printf("  - %s\n", name)
		}
	}

	fmt.Println("\nStorage Providers:")
	storageProviders := core.ListStorages()
	if len(storageProviders) == 0 {
		fmt.Println("  (none registered)")
	} else {
		for _, name := range storageProviders {
			fmt.Printf("  - %s\n", name)
		}
	}

	fmt.Println("\nUsage:")
	fmt.Println("  # Disk storage (default)")
	fmt.Println("  goarchive backup --db-host localhost [--storage-path /path/to/backups]")
	fmt.Println("\n  # S3 storage")
	fmt.Println("  goarchive backup --storage-type s3 --storage-bucket my-backups --db-host localhost")
	fmt.Println("\nExamples:")
	fmt.Println("  goarchive backup --db-type postgres --storage-type disk --db-host localhost")
	fmt.Println("  goarchive backup --db-type postgres --storage-type s3 --storage-bucket my-backups")
}
