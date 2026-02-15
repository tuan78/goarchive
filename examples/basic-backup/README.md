# Basic Backup Example

A simple example demonstrating how to use GoArchive to backup a PostgreSQL database to S3.

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

1. Run the example:

```bash
go run examples/basic-backup/main.go
```

Or using the config loader:

```bash
go run examples/using-env-config/main.go
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
