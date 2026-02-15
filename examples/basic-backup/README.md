# Basic Backup Example

A simple example demonstrating how to use GoArchive to backup a PostgreSQL database to S3.

## Setup

This example has its own `go.mod` file and can be run independently:

```bash
cd examples/basic-backup
go mod tidy  # Download dependencies
```

## Usage

1. Set environment variables:

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USERNAME=postgres
export DB_PASSWORD=secret
export DB_DATABASE=myapp

export STORAGE_BUCKET=my-backups
export STORAGE_REGION=us-west-2
export STORAGE_ACCESS_KEY=your-access-key
export STORAGE_SECRET_KEY=your-secret-key
```

2. Run the example:

```bash
# From examples/basic-backup directory
go run main.go

# Or build and run
go build -o basic-backup .
./basic-backup

# Or from the root of the project
go run examples/basic-backup/main.go
```

## What It Does

1. Creates a PostgreSQL database provider
2. Creates an S3 storage provider
3. Creates a backup service
4. Executes a backup
5. Lists all available backups

## Learn More

- See [EXTENDING.md](../../EXTENDING.md) for creating custom plugins
- Check other examples for advanced usage
