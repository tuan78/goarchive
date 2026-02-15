# Extending GoArchive

This guide explains how to add support for new databases and storage providers.

## Adding a New Database Provider

Let's walk through adding MySQL support as an example:

### 1. Create the Provider Package

Create a new directory: `database/mysql/`

### 2. Implement the Interface

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

type Provider struct {
    config *core.DatabaseConfig
    // Add MySQL-specific connection
}

func New(config *core.DatabaseConfig) (*Provider, error) {
    // Initialize MySQL connection
    return &Provider{config: config}, nil
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

func (p *Provider) GetMetadata() (*core.DatabaseMetadata, error) {
    // Query MySQL for version and size
    return &core.DatabaseMetadata{
        Type:    "mysql",
        Version: "8.0", // Query from SELECT VERSION()
        Size:    0,     // Calculate from information_schema
        Name:    p.config.Database,
    }, nil
}

func (p *Provider) Close() error {
    // Close MySQL connection
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

### 3. Update the Main Application

In `cmd/goarchive/main.go`, add the new provider:

```go
import (
    // ... existing imports
    "goarchive/database/mysql"
)

// In main():
switch config.Database.Type {
case "postgres":
    dbProvider, err = postgres.New(&config.Database)
case "mysql":
    dbProvider, err = mysql.New(&config.Database)
default:
    log.Fatalf("Unsupported database type: %s", config.Database.Type)
}
```

### 4. Update Dockerfile

Add MySQL client tools to the Dockerfile:

```dockerfile
# In the runtime stage
FROM postgres:16-alpine
RUN apk --no-cache add ca-certificates mysql-client
# ... rest of Dockerfile
```

## Adding a New Storage Provider

Let's walk through adding Azure Blob Storage support:

### 1. Create the Provider Package

Create a new directory: `storage/azure/`

### 2. Implement the Interface

Create `storage/azure/azure.go`:

```go
package azure

import (
    "context"
    "fmt"
    "io"

    "goarchive/core"

    "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

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
        return nil, err
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

    _, err := blobClient.Upload(ctx, reader, nil)
    return err
}

func (p *Provider) List(ctx context.Context) ([]*core.BackupMetadata, error) {
    // List blobs from container
    return nil, nil
}

func (p *Provider) Download(ctx context.Context, backupID string) (io.ReadCloser, error) {
    // Download from Azure Blob Storage
    return nil, nil
}

func (p *Provider) Delete(ctx context.Context, backupID string) error {
    // Delete from Azure Blob Storage
    return nil
}

func (p *Provider) getBlobName(metadata *core.BackupMetadata) string {
    return fmt.Sprintf("%s/%s_%s.dump",
        p.config.Prefix,
        metadata.DatabaseName,
        metadata.Timestamp.Format("20060102-150405"),
    )
}
```

### 3. Update StorageConfig

In `core/config.go`, add Azure-specific configuration:

```go
type StorageConfig struct {
    Type      string

    // S3 fields
    Bucket    string
    Region    string
    AccessKey string
    SecretKey string
    Prefix    string

    // Azure fields
    AccountName string
    AccountKey  string
    Container   string
}

// Update LoadConfigFromEnv() to include:
Storage: StorageConfig{
    Type:        getEnv("STORAGE_TYPE", "s3"),
    // S3
    Bucket:      getEnv("STORAGE_BUCKET", ""),
    Region:      getEnv("STORAGE_REGION", "us-east-1"),
    AccessKey:   getEnv("STORAGE_ACCESS_KEY", ""),
    SecretKey:   getEnv("STORAGE_SECRET_KEY", ""),
    Prefix:      getEnv("STORAGE_PREFIX", "backups/"),
    // Azure
    AccountName: getEnv("AZURE_STORAGE_ACCOUNT", ""),
    AccountKey:  getEnv("AZURE_STORAGE_KEY", ""),
    Container:   getEnv("AZURE_CONTAINER", "backups"),
},
```

### 4. Update the Main Application

In `cmd/goarchive/main.go`:

```go
import (
    // ... existing imports
    "goarchive/storage/azure"
)

// In main():
switch config.Storage.Type {
case "s3":
    storageProvider, err = s3.New(ctx, &config.Storage)
case "azure":
    storageProvider, err = azure.New(ctx, &config.Storage)
default:
    log.Fatalf("Unsupported storage type: %s", config.Storage.Type)
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

## Best Practices

1. **Error Handling**: Always wrap errors with context using `fmt.Errorf("context: %w", err)`
2. **Resource Cleanup**: Implement proper Close() methods and use defer
3. **Context Support**: Respect context cancellation in long-running operations
4. **Logging**: Add informative log messages for debugging
5. **Configuration Validation**: Validate provider-specific config in New()
6. **Documentation**: Add godoc comments for exported types and functions
7. **Testing**: Write both unit and integration tests

## Example: Full MongoDB Provider

See `examples/mongodb/` directory for a complete example of adding MongoDB support.

## Questions?

Open an issue on GitHub or submit a pull request with your new provider!
