# Plugin Templates

This directory contains template files for creating custom database and storage provider plugins for GoArchive.

## Templates

- **`database_provider.go.template`** - Template for creating a database provider plugin
- **`storage_provider.go.template`** - Template for creating a storage provider plugin

## Usage

1. Copy the appropriate template file to your plugin package
2. Replace placeholder names (`yourdb`, `yourstorage`) with your actual names
3. Implement the required interface methods
4. Import your plugin package to register it

## Example: Creating a MySQL Plugin

```bash
# Create package directory
mkdir -p database/mysql

# Copy template
cp _templates/database_provider.go.template database/mysql/mysql.go

# Edit the file and replace:
# - "yourdb" → "mysql"
# - Implement the methods for MySQL
```

## Complete Guide

See [EXTENDING.md](../EXTENDING.md) for a complete guide to creating plugins.

## Quick Reference

### Database Provider Interface

```go
type DatabaseProvider interface {
    Backup(ctx context.Context) (io.ReadCloser, error)
    Restore(ctx context.Context, reader io.Reader) error
    GetMetadata() (*DatabaseMetadata, error)
    Close() error
}
```

### Storage Provider Interface

```go
type StorageProvider interface {
    Upload(ctx context.Context, reader io.Reader, metadata *BackupMetadata) error
    List(ctx context.Context) ([]*BackupMetadata, error)
    Download(ctx context.Context, backupID string) (io.ReadCloser, error)
    Delete(ctx context.Context, backupID string) error
}
```

## Registration

Both provider types auto-register using `init()`:

```go
func init() {
    core.RegisterDatabase("mysql", NewMySQLProvider)
    // or
    core.RegisterStorage("gcs", NewGCSProvider)
}
```
