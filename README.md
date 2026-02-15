# GoArchive

> A flexible, plugin-based backup and restore solution for Go applications

[![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

GoArchive makes database backups simple and extensible. Built with a plugin architecture, it supports multiple databases and cloud storage providers out of the box, and makes it easy for the community to add more.

**🎯 Perfect for:**

- Production database backup automation
- Multi-database environments
- Projects requiring specific database/storage combinations
- Teams building custom backup solutions

## Features

- **🔌 Plugin Architecture**: Auto-registering plugins for databases and storage
- **📦 Use as Library**: Import core + plugins in your own projects
- **🗄️ Database Support**: PostgreSQL (more via plugins)
- **☁️ Cloud Storage**: AWS S3 and S3-compatible storage (more via plugins)
- **🔄 Backup & Restore**: Full backup and restoration support
- **🏷️ Metadata Tracking**: Automatic checksums and backup metadata
- **🐳 Docker Ready**: Containerized deployment
- **🧩 Easy to Extend**: Simple interface-based plugin system

## Installation

GoArchive uses **Go submodules** - each provider is a separate module with its own dependencies.

```bash
# Core library
go get goarchive

# Add providers you need
go get goarchive/database/postgres
go get goarchive/storage/s3
```

## Architecture

GoArchive follows a **modular multi-module architecture**:

```
goarchive/                    # Core library (no provider dependencies)
├── core/                     # Core interfaces and logic
├── cmd/goarchive/            # CLI application (separate module)
│   └── go.mod               # Imports core + selected providers
├── database/
│   └── postgres/            # PostgreSQL provider (separate module)
│       └── go.mod           # Only imports pgx
└── storage/
    ├── disk/                # Disk provider (separate module, no deps)
    │   └── go.mod
    └── s3/                  # S3 provider (separate module)
        └── go.mod           # Only imports AWS SDK
```

**Benefits:**

- ✅ Import only what you need - no forced dependencies
- ✅ Each provider is independently versioned
- ✅ Add custom providers without modifying core
- ✅ CLI and library use the same codebase

## Quick Start

### Running the CLI Tool

The `goarchive` CLI supports subcommands with flags or environment variables.

**Quick Usage:**

```bash
# Show help
goarchive help

# Show available database and storage providers
goarchive providers

# Create backup to local disk (default)
goarchive backup --db-host localhost --db-name mydb

# Backup to custom disk location
goarchive backup --db-host localhost --db-name mydb --storage-path /var/backups

# Backup to S3
goarchive backup --db-host localhost --db-name mydb --storage-type s3 --storage-bucket my-backups

# List backups from disk
goarchive list

# List backups from S3
goarchive list --storage-type s3 --storage-bucket my-backups

# Show version
goarchive version
```

**Option 1: Using command-line flags (local disk storage)**

```bash
goarchive backup \
  --db-host localhost \
  --db-port 5432 \
  --db-user postgres \
  --db-password secret \
  --db-name mydb

# Or specify custom backup location
goarchive backup \
  --db-host localhost \
  --db-name mydb \
  --storage-path /var/backups/database
```

**Option 2: Using command-line flags (S3 storage)**

```bash
goarchive backup \
  --db-host localhost \
  --db-port 5432 \
  --db-user postgres \
  --db-password secret \
  --db-name mydb \
  --storage-type s3 \
  --storage-bucket my-backups \
  --storage-region us-west-2 \
  --storage-access-key your-aws-key \
  --storage-secret-key your-aws-secret
```

**Option 3: Using environment variables**

```bash
# Set environment variables for disk storage (default)
export DB_HOST=localhost
export DB_PORT=5432
export DB_USERNAME=postgres
export DB_PASSWORD=secret
export DB_DATABASE=mydb
export STORAGE_PATH=./my-backups

# Run backup
goarchive backup

# For S3 storage
export STORAGE_TYPE=s3
export STORAGE_BUCKET=my-backups
export STORAGE_REGION=us-west-2
export STORAGE_ACCESS_KEY=your-aws-key
export STORAGE_SECRET_KEY=your-aws-secret

goarchive backup
```

**Option 4: Mixed (flags override environment variables)**

```bash
export DB_HOST=localhost
export STORAGE_PATH=./backups

# Override specific values with flags
goarchive backup --db-name mydb --db-password secret
```

**Option 5: Build and install**

```bash
# Build binary
cd cmd/goarchive && go build -o ../../goarchive .

# Or use make
make build

# Install globally
cd cmd/goarchive && go install .

# Run from anywhere
goarchive backup --db-host localhost --db-name mydb
```

**Option 6: Using Docker**

```bash
# Build image
docker build -t goarchive:latest .

# Run backup to local disk (mount volume for persistence)
docker run --rm \
  -v $(pwd)/backups:/root/backups \
  -e DB_HOST=your-db-host \
  -e DB_PASSWORD=secret \
  -e DB_DATABASE=mydb \
  goarchive:latest backup

# Run backup to S3
docker run --rm \
  -e DB_HOST=your-db-host \
  -e DB_PASSWORD=secret \
  -e DB_DATABASE=mydb \
  -e STORAGE_TYPE=s3 \
  -e STORAGE_BUCKET=my-backups \
  goarchive:latest backup

# List backups from disk
docker run --rm \
  -v $(pwd)/backups:/root/backups \
  goarchive:latest list

# List backups from S3
docker run --rm \
  -e STORAGE_TYPE=s3 \
  -e STORAGE_BUCKET=my-backups \
  goarchive:latest list
```

**Option 7: Kubernetes CronJob**

See the [Production Deployment](#production-deployment) section below for complete K8s examples.

### Complete Example with Output

```bash
# Using command-line flags
goarchive backup \
  --db-host localhost \
  --db-port 5432 \
  --db-user postgres \
  --db-password secret \
  --db-name myapp \
  --storage-bucket my-backups \
  --storage-region us-west-2 \
  --storage-access-key AKIAIOSFODNN7EXAMPLE \
  --storage-secret-key wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY \
  --storage-prefix prod-backups/

# Or using environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USERNAME=postgres
export DB_PASSWORD=secret
export DB_DATABASE=myapp
export STORAGE_BUCKET=my-backups
export STORAGE_REGION=us-west-2
export STORAGE_ACCESS_KEY=AKIAIOSFODNN7EXAMPLE
export STORAGE_SECRET_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
export STORAGE_PREFIX=prod-backups/

goarchive backup
```

**Expected Output:**

```
2026/02/15 10:30:15 Starting goarchive...
2026/02/15 10:30:15 Starting backup process...

=== Backup Completed Successfully ===
Backup ID:       myapp_postgres_20260215-103020.dump
Database:        myapp (postgres)
Timestamp:       2026-02-15T10:30:20Z
Size:            45678901 bytes (43.55 MB)
Checksum (MD5):  a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
====================================

2026/02/15 10:30:20 Backup completed successfully
```

### As a Library

```go
package main

import (
    "context"
    "log"

    "goarchive/core"

    // Import providers you need
    _ "goarchive/database/postgres"
    _ "goarchive/storage/s3"
)

func main() {
    // Create providers using the registry
    db, _ := core.GetDatabase("postgres", &core.DatabaseConfig{
        Host: "localhost",
        Port: 5432,
        Username: "postgres",
        Password: "secret",
        Database: "mydb",
    })
    defer db.Close()

    storage, _ := core.GetStorage(context.Background(), "s3", &core.StorageConfig{
        Type:   "s3",
        Bucket: "my-backups",
        Region: "us-west-2",
    })

    // Create backup service and execute
    service := core.NewBackupService(db, storage)
    metadata, _ := service.Execute(context.Background())

    log.Printf("Backup completed: %s", metadata.ID)
}
```

See [examples/](examples/) for complete working programs.

## Architecture

GoArchive uses a **plugin-based architecture** with automatic registration:

```text
goarchive/
├── core/              # Core interfaces, registry, and backup service
│   ├── backup.go      # DatabaseProvider & StorageProvider interfaces
│   ├── registry.go    # Plugin registration system
│   └── config.go      # Configuration structures
├── database/          # Database provider plugins
│   └── postgres/      # PostgreSQL plugin (auto-registers via init)
├── storage/           # Storage provider plugins
│   └── s3/            # AWS S3 plugin (auto-registers via init)
├── cmd/goarchive/     # CLI application
└── examples/          # Usage examples and plugin templates
```

### How It Works

1. **Plugins register themselves** via `init()` functions
2. **Import plugins** as blank imports: `_ "goarchive/database/postgres"`
3. **Get providers** from registry: `core.GetDatabase("postgres", config)`
4. **Plugins are decoupled** from the core - add/remove as needed

## Docker Deployment

GoArchive supports two deployment modes:

1. **Scheduled Backups** (Recommended): Runs continuously with automatic backups on a schedule
2. **One-time Backups**: Runs a single backup and exits

See [docs/DOCKER_COMPOSE.md](docs/DOCKER_COMPOSE.md) for detailed deployment guide.

### Quick Start: Scheduled Backups

Run continuous backups on a schedule (default: every 6 hours):

```bash
# Start infrastructure and scheduled backup service
docker-compose up -d

# View logs
docker-compose logs -f goarchive-scheduled

# The default Dockerfile creates /root/backups directory for disk storage
# Backups are stored in the backup-data Docker volume
```

Customize the schedule with environment variables:

```yaml
environment:
  BACKUP_SCHEDULE: "0 2 * * *" # Daily at 2 AM (cron expression)
  # Or use interval: BACKUP_INTERVAL: "3600"  # Every 1 hour (in seconds)
```

### Quick Start: One-time Backup

For testing or manual backups:

```bash
# One-time backup to disk
docker-compose --profile oneshot run --rm goarchive-disk

# One-time backup to S3
docker-compose --profile oneshot run --rm goarchive
```

### Using Docker Compose for Testing

1. **Start the test environment**:

   ```bash
   docker-compose up -d postgres localstack
   ```

2. **Wait for services to be ready** (about 10 seconds)

3. **Create the S3 bucket in LocalStack**:

   ```bash
   docker-compose run --rm goarchive sh -c "
     aws --endpoint-url=http://localstack:4566 s3 mb s3://backups --region us-east-1
   "
   ```

4. **Run the backup**:

   ```bash
   docker-compose run --rm goarchive
   ```

### Using Docker

1. **Build the image**:

   ```bash
   # For scheduled backups
   docker build -f Dockerfile.scheduler -t goarchive:scheduler .

   # For one-time backups
   docker build -t goarchive .
   ```

2. **Run scheduled backups**:

   ```bash
   docker run -d \
     --name goarchive-scheduler \
     --restart unless-stopped \
     -e DB_HOST=your-postgres-host \
     -e DB_PORT=5432 \
     -e DB_USERNAME=postgres \
     -e DB_PASSWORD=your-password \
     -e DB_DATABASE=your-database \
     -e STORAGE_TYPE=disk \
     -e STORAGE_PATH=/root/backups \
     -e BACKUP_SCHEDULE="0 */6 * * *" \
     -v /host/backups:/root/backups \
     goarchive:scheduler
   ```

3. **Run one-time backup**:

   ```bash
   docker run --rm \
     -e DB_HOST=your-postgres-host \
     -e DB_PORT=5432 \
     -e DB_USERNAME=postgres \
     -e DB_PASSWORD=your-password \
     -e DB_DATABASE=your-database \
     -e STORAGE_BUCKET=your-s3-bucket \
     -e STORAGE_REGION=us-east-1 \
     -e STORAGE_ACCESS_KEY=your-access-key \
     -e STORAGE_SECRET_KEY=your-secret-key \
     goarchive backup
   ```

> **Note**: The Dockerfile creates `/root/backups` directory by default for disk storage. This is mounted as a volume for persistence.

### Local Development

1. **Install dependencies**:

   ```bash
   go mod download
   ```

2. **Set environment variables**:

   ```bash
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_USERNAME=postgres
   export DB_PASSWORD=yourpassword
   export DB_DATABASE=yourdatabase
   export STORAGE_BUCKET=your-bucket
   export STORAGE_REGION=us-east-1
   export STORAGE_ACCESS_KEY=your-key
   export STORAGE_SECRET_KEY=your-secret
   ```

3. **Run the application**:

   ```bash
   cd cmd/goarchive && go run main.go backup

   # Or use make
   make run
   ```

## Configuration

All configuration is done through environment variables:

### Database Configuration

| Variable      | Description                                       | Default     |
| ------------- | ------------------------------------------------- | ----------- |
| `DB_TYPE`     | Database type (run `goarchive providers` to list) | `postgres`  |
| `DB_HOST`     | Database host                                     | `localhost` |
| `DB_PORT`     | Database port                                     | `5432`      |
| `DB_USERNAME` | Database username                                 | `postgres`  |
| `DB_PASSWORD` | Database password                                 | -           |
| `DB_DATABASE` | Database name                                     | `postgres`  |
| `DB_SSLMODE`  | SSL mode (`disable`, `require`, `verify-full`)    | `disable`   |

### Storage Configuration

| Variable             | Description                                      | Default     |
| -------------------- | ------------------------------------------------ | ----------- |
| `STORAGE_TYPE`       | Storage type (run `goarchive providers` to list) | `disk`      |
| `STORAGE_PATH`       | Local directory for backups (disk storage)       | `./backups` |
| `STORAGE_BUCKET`     | S3 bucket name (S3 storage)                      | -           |
| `STORAGE_REGION`     | AWS region (S3 storage)                          | `us-east-1` |
| `STORAGE_ACCESS_KEY` | AWS access key (S3, optional if using IAM)       | -           |
| `STORAGE_SECRET_KEY` | AWS secret key (S3, optional if using IAM)       | -           |
| `STORAGE_PREFIX`     | S3 prefix for backups (S3 storage)               | `backups/`  |
| `AWS_ENDPOINT_URL`   | Custom S3 endpoint (for LocalStack/MinIO)        | -           |

## Available Providers

To see all available providers in your installation, run:

```bash
goarchive providers
```

### Selecting Providers

Use the `--db-type` and `--storage-type` flags (or `DB_TYPE` and `STORAGE_TYPE` environment variables) to select which providers to use:

```bash
# Using command-line flags
goarchive backup --db-type postgres --storage-type s3 [other options]

# Using environment variables
export DB_TYPE=postgres
export STORAGE_TYPE=s3
goarchive backup
```

### Database Providers

- **postgres** - PostgreSQL databases (using `pg_dump`/`pg_restore`)

### Storage Providers

- **disk** - Local disk storage (default, no additional dependencies)
- **s3** - Amazon S3 and S3-compatible storage (MinIO, LocalStack, etc.)

> **Note:** Additional providers can be added as separate Go modules. See [EXTENDING.md](EXTENDING.md) for details on creating custom database or storage providers.

### Community Plugins

Want to add your plugin here? Submit a PR or create an issue!

## Troubleshooting

### Common Issues

**Error: "could not import goarchive/database/postgres"**

```bash
# Solution: Install the provider submodule
go get goarchive/database/postgres
go get goarchive/storage/s3
go mod tidy
```

**Error: "database provider 'postgres' not registered"**

```go
// Solution: Make sure you import the provider with blank import
import _ "goarchive/database/postgres"
```

**Error: "failed to connect to PostgreSQL"**

- Check `DB_HOST`, `DB_PORT`, `DB_USERNAME`, `DB_PASSWORD` are correct
- Verify the database is accessible from your network
- Check if SSL is required: set `DB_SSLMODE=require`

**Error: "failed to upload to S3"**

- Verify `STORAGE_BUCKET` exists and you have write permissions
- Check `STORAGE_ACCESS_KEY` and `STORAGE_SECRET_KEY` are valid
- Ensure `STORAGE_REGION` matches your bucket's region

**Docker build fails: "requires go >= 1.24.0"**

- Update Dockerfile to use a newer Go version:
  ```dockerfile
  FROM golang:1.24-alpine AS builder
  ```

### Testing Locally

```bash
# Test with LocalStack (S3) and local PostgreSQL
docker-compose up -d postgres localstack

# Wait for services to start
sleep 10

# Create S3 bucket in LocalStack
docker-compose run --rm goarchive sh -c \
  "aws --endpoint-url=http://localstack:4566 s3 mb s3://backups --region us-east-1"

# Run backup
docker-compose run --rm goarchive
```

## Extending GoArchive

### Creating a Plugin

It's easy to add support for new databases or storage backends:

1. **Implement the interface** (`DatabaseProvider` or `StorageProvider`)
2. **Register in init()** function
3. **Import the plugin** in your application

**Example Database Plugin:**

```go
package mysql

import "goarchive/core"

func init() {
    // Auto-register when imported
    core.RegisterDatabase("mysql", func(config *core.DatabaseConfig) (core.DatabaseProvider, error) {
        return New(config)
    })
}

type Provider struct { /* ... */ }
func (p *Provider) Backup(ctx context.Context) (io.ReadCloser, error) { /* ... */ }
func (p *Provider) Restore(ctx context.Context, reader io.Reader) error { /* ... */ }
func (p *Provider) GetMetadata() (*core.DatabaseMetadata, error) { /* ... */ }
func (p *Provider) Close() error { /* ... */ }
```

**Usage:**

```go
import _ "goarchive/database/mysql"  // Auto-registers

db, _ := core.GetDatabase("mysql", config)
```

**Complete guides:**

- 📖 [EXTENDING.md](EXTENDING.md) - Detailed plugin development guide
- 📁 [\_templates/](_templates/) - Plugin templates
- 📁 [examples/](examples/) - Working example programs

### Publishing Plugins

Plugins can be **published as separate Go modules**:

```go
// github.com/youruser/goarchive-mongodb/go.mod
module github.com/youruser/goarchive-mongodb

go 1.24.0

require (
    goarchive v0.1.0
    go.mongodb.org/mongo-driver v1.13.0
)
```

Users install and use it:

```bash
go get github.com/youruser/goarchive-mongodb
```

```go
import _ "github.com/youruser/goarchive-mongodb"
```

## Production Deployment

### Building for Production

**Build Docker Image:**

```bash
# Build the image
docker build -t your-registry/goarchive:1.0.0 .

# Push to registry
docker push your-registry/goarchive:1.0.0
```

**Or build native binary:**

```bash
# Build for Linux (for K8s/containers)
cd cmd/goarchive && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../../goarchive .

# Build for your platform
cd cmd/goarchive && go build -o ../../goarchive .

# Or use make
make build
```

### Kubernetes CronJob

**Step 1: Create Secret for credentials**

```bash
kubectl create secret generic goarchive-secrets \
  --from-literal=db-password='your-db-password' \
  --from-literal=storage-access-key='your-aws-access-key' \
  --from-literal=storage-secret-key='your-aws-secret-key'
```

**Step 2: Apply CronJob manifest**

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: database-backup
spec:
  schedule: "0 2 * * *" # Daily at 2 AM UTC
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 3
  jobTemplate:
    spec:
      template:
        spec:
          restartPolicy: OnFailure
          containers:
            - name: goarchive
              image: your-registry/goarchive:1.0.0
              command: ["goarchive", "backup"]
              env:
                # Database configuration
                - name: DB_TYPE
                  value: "postgres"
                - name: DB_HOST
                  value: "postgres-service.default.svc.cluster.local"
                - name: DB_PORT
                  value: "5432"
                - name: DB_DATABASE
                  value: "myapp"
                - name: DB_USERNAME
                  value: "postgres"
                - name: DB_PASSWORD
                  valueFrom:
                    secretKeyRef:
                      name: goarchive-secrets
                      key: db-password
                - name: DB_SSLMODE
                  value: "require"

                # Storage configuration
                - name: STORAGE_TYPE
                  value: "s3"
                - name: STORAGE_BUCKET
                  value: "my-backup-bucket"
                - name: STORAGE_REGION
                  value: "us-east-1"
                - name: STORAGE_PREFIX
                  value: "prod-backups/"
                - name: STORAGE_ACCESS_KEY
                  valueFrom:
                    secretKeyRef:
                      name: goarchive-secrets
                      key: storage-access-key
                - name: STORAGE_SECRET_KEY
                  valueFrom:
                    secretKeyRef:
                      name: goarchive-secrets
                      key: storage-secret-key

              resources:
                requests:
                  memory: "128Mi"
                  cpu: "100m"
                limits:
                  memory: "512Mi"
                  cpu: "500m"
```

**Step 3: Verify the CronJob**

```bash
# Check CronJob status
kubectl get cronjob database-backup

# View recent jobs
kubectl get jobs

# Check logs from a job
kubectl logs job/database-backup-<timestamp>

# Manually trigger a job (for testing)
kubectl create job backup-manual --from=cronjob/database-backup
```

**Alternative: Using Disk Storage with PersistentVolumeClaim**

```yaml
# PersistentVolumeClaim for backup storage
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: backup-storage
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
---
# CronJob using disk storage
apiVersion: batch/v1
kind: CronJob
metadata:
  name: database-backup-disk
spec:
  schedule: "0 2 * * *"
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 3
  jobTemplate:
    spec:
      template:
        spec:
          restartPolicy: OnFailure
          containers:
            - name: goarchive
              image: your-registry/goarchive:1.0.0
              command: ["goarchive", "backup"]
              env:
                # Database configuration
                - name: DB_HOST
                  value: "postgres-service.default.svc.cluster.local"
                - name: DB_PORT
                  value: "5432"
                - name: DB_DATABASE
                  value: "myapp"
                - name: DB_USERNAME
                  value: "postgres"
                - name: DB_PASSWORD
                  valueFrom:
                    secretKeyRef:
                      name: goarchive-secrets
                      key: db-password
                - name: DB_SSLMODE
                  value: "require"

                # Storage configuration (disk)
                - name: STORAGE_TYPE
                  value: "disk"
                - name: STORAGE_PATH
                  value: "/backups"

              volumeMounts:
                - name: backup-storage
                  mountPath: /backups

              resources:
                requests:
                  memory: "128Mi"
                  cpu: "100m"
                limits:
                  memory: "512Mi"
                  cpu: "500m"

          volumes:
            - name: backup-storage
              persistentVolumeClaim:
                claimName: backup-storage
```

### AWS ECS Scheduled Task

Configure the task with the appropriate environment variables and IAM role with permissions for S3 and RDS access.

## Security Best Practices

1. **Never hardcode credentials** - use environment variables or secrets management
2. **Use IAM roles** when running in AWS (no need for access keys)
3. **Enable SSL/TLS** for database connections in production
4. **Encrypt backups** at rest in S3 (use S3 bucket encryption)
5. **Use least privilege** - grant only necessary permissions
6. **Rotate credentials** regularly

## Contributing

We welcome contributions! Whether you're:

- 🐛 Reporting bugs
- 💡 Suggesting features
- 🔌 **Creating new plugins** (database or storage providers)
- 📚 Improving documentation
- 🧪 Adding tests

Please see our [Contributing Guide](CONTRIBUTING.md) and [Plugin Development Guide](EXTENDING.md).

### Quick Plugin Contribution Guide

1. Use templates from `_templates/` directory
2. Implement the required interface (`DatabaseProvider` or `StorageProvider`)
3. Add auto-registration in `init()` function
4. Include tests and documentation
5. Submit a PR!

See [EXTENDING.md](EXTENDING.md) for detailed instructions.

## Roadmap

Community-requested features:

- [ ] MySQL/MariaDB provider
- [ ] MongoDB provider
- [ ] SQLite provider
- [ ] Azure Blob Storage
- [ ] Google Cloud Storage
- [ ] Backup encryption before upload
- [ ] Backup compression options
- [ ] Backup retention policies
- [ ] Email/Slack notifications
- [ ] Prometheus metrics

**Want to contribute?** Pick an item and [open an issue](https://github.com/tuan78/goarchive/issues) to discuss implementation!

## License

MIT License - feel free to use this in your projects!

## Support

- 📖 [Documentation](README.md)
- 🔌 [Plugin Development Guide](EXTENDING.md)
- 🤝 [Contributing Guide](CONTRIBUTING.md)
- 🐛 [Report Issues](https://github.com/tuan78/goarchive/issues)
- 💬 [Discussions](https://github.com/tuan78/goarchive/discussions)
