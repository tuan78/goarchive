# Extending GoArchive

This guide explains how to add support for new databases and storage providers.

## Architecture: Go Submodules

Each provider is a separate module with its own `go.mod`:

```
goarchive/
├── go.mod              # Core library
├── core/               # Interfaces and backup logic
├── database/
│   └── postgres/
│       ├── go.mod      # pgx dependency
│       └── postgres.go
└── storage/
    └── s3/
        ├── go.mod      # AWS SDK dependency
        └── s3.go
```

**Benefits:**

- Users install only what they need
- Independent versioning per provider
- Easy to publish third-party providers

## Adding a New Database Provider

Let's walk through adding MySQL support as an example.

### 1. Create the Provider Package

Create a new directory: `database/mysql/`

### 2. Create go.mod for the Provider

Create `database/mysql/go.mod`:

```go
module goarchive/database/mysql

go 1.24.0

require (
    goarchive v0.0.0
    github.com/go-sql-driver/mysql v1.7.1
)

replace goarchive => ../../
```

### 3. Implement the Interface

Create `database/mysql/mysql.go`:

```go
package mysql

import (
    "context"
    "fmt"
    "io"
    "os/exec"

    "goarchive/core"

    "github.com/go-sql-driver/mysql"
)

// init registers the MySQL provider with the global registry
// This runs automatically when the package is imported
func init() {
    core.RegisterDatabase("mysql", func(config *core.DatabaseConfig) (core.DatabaseProvider, error) {
        return New(config)
    })
}

type Provider struct {
    config *core.DatabaseConfig
    db     *sql.DB
}

func New(config *core.DatabaseConfig) (*Provider, error) {
    // Initialize MySQL connection
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
        config.Username,
        config.Password,
        config.Host,
        config.Port,
        config.Database,
    )

    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
    }

    return &Provider{config: config, db: db}, nil
}

func (p *Provider) Backup(ctx context.Context) (io.ReadCloser, error) {
    // Use mysqldump to create backup
    cmd := exec.CommandContext(ctx, "mysqldump",
        "-h", p.config.Host,
        "-P", fmt.Sprintf("%d", p.config.Port),
        "-u", p.config.Username,
        fmt.Sprintf("-p%s", p.config.Password),
        p.config.Database,
    )

    stdout, err := cmd.StdoutPipe()
    if err != nil {
        return nil, err
    }

    if err := cmd.Start(); err != nil {
        return nil, err
    }

    // Return reader that waits for command completion
    return &backupReader{ReadCloser: stdout, cmd: cmd}, nil
}

func (p *Provider) Restore(ctx context.Context, reader io.Reader) error {
    // Use mysql client to restore
    cmd := exec.CommandContext(ctx, "mysql",
        "-h", p.config.Host,
        "-P", fmt.Sprintf("%d", p.config.Port),
        "-u", p.config.Username,
        fmt.Sprintf("-p%s", p.config.Password),
        p.config.Database,
    )

    cmd.Stdin = reader
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("failed to restore: %w (output: %s)", err, string(output))
    }
    return nil
}

func (p *Provider) GetMetadata() (*core.DatabaseMetadata, error) {
    var version string
    err := p.db.QueryRow("SELECT VERSION()").Scan(&version)
    if err != nil {
        return nil, fmt.Errorf("failed to get version: %w", err)
    }

    return &core.DatabaseMetadata{
        Type:    "mysql",
        Version: version,
        Name:    p.config.Database,
    }, nil
}

func (p *Provider) Close() error {
    if p.db != nil {
        return p.db.Close()
    }
    return nil
}

type backupReader struct {
    io.ReadCloser
    cmd *exec.Cmd
}

func (r *backupReader) Close() error {
    r.ReadCloser.Close()
    return r.cmd.Wait()
}
```

### 4. Usage

Import your provider to enable it:

```go
package main

import (
    "goarchive/core"
    _ "goarchive/database/mysql"
    _ "goarchive/storage/s3"
)

func main() {
    db, _ := core.GetDatabase("mysql", &core.DatabaseConfig{
        Host:     "localhost",
        Port:     3306,
        Username: "root",
        Password: "password",
        Database: "myapp",
    })
    defer db.Close()
    // ... use db
}
```

The `init()` function registers the provider automatically.

## Adding a New Storage Provider

Let's walk through adding Azure Blob Storage support.

### 1. Create the Provider Package

Create a new directory: `storage/azure/`

### 2. Create go.mod for the Provider

Create `storage/azure/go.mod`:

```go
module goarchive/storage/azure

go 1.24.0

require (
    goarchive v0.0.0
    github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.2.0
)

replace goarchive => ../../
```

### 3. Implement the Interface

Create `storage/azure/azure.go`:

```go
package azure

import (
    "context"
    "fmt"
    "io"
    "time"

    "goarchive/core"

    "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

// init registers the Azure provider with the global registry
func init() {
    core.RegisterStorage("azure", func(ctx context.Context, config *core.StorageConfig) (core.StorageProvider, error) {
        return New(ctx, config)
    })
}

type Provider struct {
    client *azblob.Client
    config *core.StorageConfig
}

func New(ctx context.Context, config *core.StorageConfig) (*Provider, error) {
    // Initialize Azure client
    client, err := azblob.NewClientWithSharedKey(
        fmt.Sprintf("https://%s.blob.core.windows.net/", config.AccountName),
        &azblob.SharedKeyCredential{
            AccountName: config.AccountName,
            AccountKey:  config.AccountKey,
        },
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create Azure client: %w", err)
    }

    return &Provider{
        client: client,
        config: config,
    }, nil
}

func (p *Provider) Upload(ctx context.Context, reader io.Reader, metadata *core.BackupMetadata) error {
    // Upload to Azure Blob Storage
    containerClient := p.client.ServiceClient().NewContainerClient(p.config.Container)
    blobName := p.getBlobName(metadata)
    blobClient := containerClient.NewBlockBlobClient(blobName)

    _, err := blobClient.Upload(ctx, reader, &azblob.UploadBlockBlobOptions{
        Metadata: map[string]*string{
            "database-name": &metadata.DatabaseName,
            "database-type": &metadata.DatabaseType,
            "backup-id":     &metadata.ID,
        },
    })
    return err
}

func (p *Provider) List(ctx context.Context) ([]*core.BackupMetadata, error) {
    // List blobs from container
    containerClient := p.client.ServiceClient().NewContainerClient(p.config.Container)
    pager := containerClient.NewListBlobsFlatPager(&azblob.ListBlobsFlatOptions{
        Prefix: &p.config.Prefix,
    })

    var backups []*core.BackupMetadata
    for pager.More() {
        page, err := pager.NextPage(ctx)
        if err != nil {
            return nil, fmt.Errorf("failed to list blobs: %w", err)
        }

        for _, blob := range page.Segment.BlobItems {
            backups = append(backups, &core.BackupMetadata{
                ID:        *blob.Name,
                Timestamp: *blob.Properties.LastModified,
                Size:      *blob.Properties.ContentLength,
            })
        }
    }
    return backups, nil
}

func (p *Provider) Download(ctx context.Context, backupID string) (io.ReadCloser, error) {
    // Download from Azure Blob Storage
    containerClient := p.client.ServiceClient().NewContainerClient(p.config.Container)
    blobClient := containerClient.NewBlockBlobClient(backupID)

    response, err := blobClient.DownloadStream(ctx, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to download blob: %w", err)
    }

    return response.Body, nil
}

func (p *Provider) Delete(ctx context.Context, backupID string) error {
    // Delete from Azure Blob Storage
    containerClient := p.client.ServiceClient().NewContainerClient(p.config.Container)
    blobClient := containerClient.NewBlockBlobClient(backupID)

    _, err := blobClient.Delete(ctx, nil)
    return err
}

func (p *Provider) getBlobName(metadata *core.BackupMetadata) string {
    return fmt.Sprintf("%s/%s_%s_%s.dump",
        p.config.Prefix,
        metadata.DatabaseName,
        metadata.DatabaseType,
        metadata.Timestamp.Format("20060102-150405"),
    )
}
```

### 4. Usage

```go
package main

import (
    "goarchive/core"
    _ "goarchive/database/postgres"
    _ "goarchive/storage/azure"
)

func main() {
    storage, _ := core.GetStorage(ctx, "azure", &core.StorageConfig{
        Type:        "azure",
        AccountName: "mystorageaccount",
        AccountKey:  "key",
        Container:   "backups",
        Prefix:      "db-backups/",
    })
    // ... use storage
}
```

## Testing Your Extensions

### Unit Tests

Create tests for your provider in the same package:

```go
// database/mysql/mysql_test.go
package mysql

import (
    "testing"
    "goarchive/core"
)

func TestNew(t *testing.T) {
    config := &core.DatabaseConfig{
        Host:     "localhost",
        Port:     3306,
        Username: "root",
        Database: "test",
    }

    provider, err := New(config)
    if err != nil {
        t.Fatalf("Failed to create provider: %v", err)
    }

    if provider == nil {
        t.Fatal("Provider is nil")
    }
}

func TestAutoRegistration(t *testing.T) {
    // Test that the provider auto-registered via init()
    db, err := core.GetDatabase("mysql", &core.DatabaseConfig{
        Host:     "localhost",
        Port:     3306,
        Username: "root",
        Database: "test",
    })

    if err != nil {
        t.Fatalf("Provider not registered: %v", err)
    }

    if db == nil {
        t.Fatal("Database provider is nil")
    }
}
```

### Integration Tests

Add your provider to the docker-compose.yml:

```yaml
mysql:
  image: mysql:8.0
  environment:
    MYSQL_ROOT_PASSWORD: testpass
    MYSQL_DATABASE: testdb
  ports:
    - "3306:3306"
```

## Publishing Third-Party Providers

Create a separate repository with its own module:

### 1. Create Repository Structure

```bash
github.com/youruser/goarchive-mongodb/
├── go.mod
├── mongodb.go
├── mongodb_test.go
└── README.md
```

### 2. Create go.mod

```go
module github.com/youruser/goarchive-mongodb

go 1.24.0

require (
    goarchive v0.1.0
    go.mongodb.org/mongo-driver v1.13.0
)
```

### 3. Implement with Auto-Registration

```go
package mongodb

import (
    "goarchive/core"
    "go.mongodb.org/mongo-driver/mongo"
)

func init() {
    core.RegisterDatabase("mongodb", func(config *core.DatabaseConfig) (core.DatabaseProvider, error) {
        return New(config)
    })
}

// ... implement Provider
```

### 4. Publish

```bash
git tag v0.1.0
git push origin v0.1.0
```

### 5. Users Install Your Provider

```bash
go get github.com/youruser/goarchive-mongodb
```

```go
import _ "github.com/youruser/goarchive-mongodb"

db, _ := core.GetDatabase("mongodb", config)
```

## Best Practices

1. **Auto-Registration**: Always use `init()` to register providers automatically
2. **Error Handling**: Always wrap errors with context using `fmt.Errorf("context: %w", err)`
3. **Resource Cleanup**: Implement proper `Close()` methods and use `defer`
4. **Context Support**: Respect context cancellation in long-running operations
5. **Logging**: Add informative log messages for debugging
6. **Configuration Validation**: Validate provider-specific config in `New()`
7. **Documentation**: Add godoc comments for exported types and functions
8. **Testing**: Write both unit and integration tests
9. **Submodules**: Use separate `go.mod` to minimize dependencies
10. **Versioning**: Follow semantic versioning for your provider modules

## Configuration Extensions

If your provider needs custom config fields, extend `StorageConfig` or `DatabaseConfig`:

```go
// In core/config.go - add fields for multiple providers
type StorageConfig struct {
    Type   string
    Prefix string

    // S3
    Bucket    string
    Region    string
    AccessKey string
    SecretKey string

    // Azure
    AccountName string
    AccountKey  string
    Container   string

    // GCS
    ProjectID      string
    CredentialPath string
}
```

Providers ignore fields they don't use, so there's no conflict.

## Questions?

Open an issue on GitHub or submit a pull request with your new provider!
