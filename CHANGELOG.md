# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-02-15

### Added

#### Core Features

- **Multi-module architecture** with proper dependency isolation
  - Root `go.mod`: Core library only with zero dependencies
  - `cmd/goarchive/go.mod`: CLI application with all provider dependencies
  - Provider submodules: `database/postgres`, `storage/disk`, `storage/s3`
  - Independent example modules: `examples/basic-backup`, `examples/using-env-config`

- **Disk storage provider** as default storage backend
  - Local filesystem storage with zero external dependencies
  - MD5 checksum validation for backup integrity
  - Automatic metadata file creation (.meta files)
  - Default backup directory: `./backups`

- **PostgreSQL database provider** for backups
  - Uses `pg_dump` for reliable backup generation
  - Supports all standard PostgreSQL connection options
  - SSL/TLS connection support

- **S3 storage provider** for cloud backups
  - AWS S3 compatible storage support
  - Custom endpoint support (works with LocalStack, MinIO, etc.)
  - Automatic gzip compression
  - Configurable bucket, region, and path prefix

#### CLI Enhancements

- **Natural command-line interface** with subcommands
  - `backup`: Create a new backup
  - `list`: List available backups
  - `providers`: Show available database and storage providers
  - `version`: Display version information
  - `help`: Show command help

- **Flag-based arguments** with environment variable fallback
  - Database flags: `--db-host`, `--db-port`, `--db-username`, etc.
  - Storage flags: `--storage-type`, `--storage-path`, `--storage-bucket`, etc.
  - All flags have corresponding environment variables

- **Dynamic provider listing** showing available and configured providers

#### Docker Support

- **Scheduled backups** with cron support
  - `Dockerfile.scheduler`: Image with cron daemon
  - `scripts/scheduler.sh`: Flexible scheduler supporting cron expressions or intervals
  - Environment variables: `BACKUP_SCHEDULE` (cron) or `BACKUP_INTERVAL` (seconds)
  - Automatic restart with `restart: unless-stopped` policy

- **Docker Compose configurations**
  - `oneshot` profile: Run single backup and exit
  - `scheduled` profile: Run continuous scheduled backups (default)
  - Separate services for disk and S3 storage
  - LocalStack integration for local S3 testing
  - PostgreSQL test database included

- **Default backup directory** in Docker images
  - Pre-created `/root/backups` directory
  - Declared as `VOLUME` for persistence
  - Easy host mounting support

#### Documentation

- Comprehensive README with architecture overview
- `docs/DOCKER_COMPOSE.md`: Complete Docker deployment guide
  - Production deployment examples
  - Scheduling configuration
  - Monitoring and troubleshooting
  - Backup retention strategies

- Updated example documentation
  - Independent module build instructions
  - Multiple run options (from example dir or project root)
  - Environment variable configuration

### Changed

- **Removed hardcoded environment variables** from Dockerfile
  - All configuration now provided at runtime
  - Better security and flexibility

- **Default storage** changed from S3 to disk
  - No cloud setup required for basic usage
  - S3 still fully supported as option

- **Separated CLI module** from core library
  - Library users don't download unused provider dependencies
  - Cleaner dependency tree

### Developer Experience

- **Make targets** for common tasks
  - `make build`: Build the CLI binary
  - `make tidy`: Tidy all module dependencies
  - `make test`: Run all tests
  - `make docker-build`: Build standard Docker image
  - `make docker-build-scheduler`: Build scheduler Docker image

- **Examples as standalone modules**
  - Each example has its own `go.mod`
  - Can be built independently
  - Include `.gitignore` for compiled binaries

### Technical Details

- Go version: 1.24.0
- PostgreSQL driver: pgx v5.8.0
- AWS SDK: v2.41.1
- Docker base images: golang:1.24-alpine, postgres:16-alpine

[0.1.0]: https://github.com/yourusername/goarchive/releases/tag/v0.1.0
