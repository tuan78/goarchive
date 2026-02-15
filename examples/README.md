# GoArchive Examples

This directory contains working example programs demonstrating how to use GoArchive.

## Independent Modules

Each example has its own `go.mod` file, making them independently buildable and runnable. This allows you to:

- Build and run examples without needing to understand the entire project structure
- Use examples as starting templates for your own projects
- Test the library integration in isolation

## Examples

### [basic-backup](basic-backup/)

A simple example showing how to create a backup using PostgreSQL and S3.

**Run:**

```bash
cd examples/basic-backup
go mod tidy  # First time only
go run main.go

# Or from the project root
go run examples/basic-backup/main.go
```

### [using-env-config](using-env-config/)

Demonstrates how to use the built-in environment variable configuration loader.

**Run:**

```bash
cd examples/using-env-config
go mod tidy  # First time only

export DB_TYPE=postgres
export DB_HOST=localhost
# ... set other env vars
go run main.go

# Or from the project root
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

Each example is an independent Go module. You can run them in two ways:

**From the example directory:**

```bash
cd examples/basic-backup
go mod tidy  # First time only
go run main.go

# Or build them
go build -o basic-backup .
./basic-backup
```

**From the project root:**

```bash
go run examples/basic-backup/main.go
```

The examples use Go module `replace` directives to depend on local goarchive modules during development.

## Configuration

Most examples expect environment variables to be set. See the individual example README files for specific requirements.

## Learn More

- 📖 [Main README](../README.md) - Project overview
- 📖 [EXTENDING.md](../EXTENDING.md) - Plugin development guide
- 🔧 [\_templates/](_templates/) - Plugin code templates
