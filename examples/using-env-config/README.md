# Using Environment Config

This example shows how to use the built-in configuration loader that reads from environment variables.

## Setup

This example has its own `go.mod` file and can be run independently:

```bash
cd examples/using-env-config
go mod tidy  # Download dependencies
```

## Usage

```bash
# Set environment variables
export DB_TYPE=postgres
export DB_HOST=localhost
export DB_PORT=5432
export DB_USERNAME=postgres
export DB_PASSWORD=secret
export DB_DATABASE=myapp

export STORAGE_TYPE=s3
export STORAGE_BUCKET=my-backups
export STORAGE_REGION=us-west-2
export STORAGE_ACCESS_KEY=your-key
export STORAGE_SECRET_KEY=your-secret

# Run the example from examples/using-env-config directory
go run main.go

# Or build and run
go build -o using-env-config .
./using-env-config

# Or from the root of the project
go run examples/using-env-config/main.go
```

## Supported Environment Variables

See [README.md](../../README.md#configuration) for the complete list of configuration options.

## Benefits

- No hardcoded credentials
- Easy to use with Docker, Kubernetes, etc.
- Consistent with the main CLI tool
- Supports defaults for optional values
