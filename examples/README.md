# GoArchive Examples

This directory contains working example programs demonstrating how to use GoArchive.

## Examples

### [basic-backup](basic-backup/)

A simple example showing how to create a backup using PostgreSQL and S3.

**Run:**

```bash
go run examples/basic-backup/main.go
```

### [using-env-config](using-env-config/)

Demonstrates how to use the built-in environment variable configuration loader.

**Run:**

```bash
export DB_TYPE=postgres
export DB_HOST=localhost
# ... set other env vars
go run examples/using-env-config/main.go
```

## Example Test Functions

The core package also includes Example functions for godoc:

```bash
# View examples in documentation
go doc -all goarchive/core | grep Example

# Run example tests
go test goarchive/core -v -run Example
```

## Creating Your Own Plugin

See the [\_templates/](_templates/) directory for plugin templates and [EXTENDING.md](../EXTENDING.md) for a complete guide.

## Running Examples

All examples are standard Go programs that import the goarchive module:

```bash
# Run any example from the project root
go run examples/basic-backup/main.go

# Or build them
go build -o backup examples/basic-backup/main.go
./backup
```

## Configuration

Most examples expect environment variables to be set. See the individual example README files for specific requirements.

## Learn More

- 📖 [Main README](../README.md) - Project overview
- 📖 [EXTENDING.md](../EXTENDING.md) - Plugin development guide
- 🔧 [\_templates/](_templates/) - Plugin code templates
