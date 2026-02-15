# Contributing to GoArchive

Thank you for your interest in contributing to GoArchive! This document provides guidelines for contributing to the project.

## How to Contribute

### Reporting Bugs

If you find a bug, please create an issue with:

- A clear, descriptive title
- Steps to reproduce the issue
- Expected vs actual behavior
- Your environment (Go version, OS, database/storage versions)
- Relevant logs or error messages

### Suggesting Features

We welcome feature suggestions! Please create an issue with:

- A clear description of the feature
- Use cases and benefits
- Any implementation ideas you have

### Contributing Code

We love pull requests! Here's how to contribute:

1. **Fork the repository** and create a branch from `main`
2. **Make your changes** following our coding standards
3. **Add tests** for new functionality
4. **Update documentation** as needed
5. **Ensure tests pass**: `go test ./...`
6. **Submit a pull request**

## Development Setup

### Prerequisites

- Go 1.24 or later
- Docker and Docker Compose (for integration tests)
- PostgreSQL client tools (`pg_dump`, `pg_restore`)

### Getting Started

```bash
# Clone your fork
git clone https://github.com/tuan78/goarchive
cd goarchive

# Install dependencies
go mod download

# Run tests
go test ./...

# Run integration tests with Docker
docker-compose up -d postgres localstack
go test ./... -tags=integration
```

## Creating a Plugin

The best way to contribute is by adding new database or storage providers! See our [plugin development guide](EXTENDING.md) for detailed instructions.

### Database Provider Checklist

- [ ] Create package under `database/yourdb/`
- [ ] Implement `core.DatabaseProvider` interface
- [ ] Add auto-registration in `init()` function
- [ ] Include unit tests
- [ ] Add integration tests (Docker Compose service)
- [ ] Update documentation (README.md, EXTENDING.md)
- [ ] Add example usage

### Storage Provider Checklist

- [ ] Create package under `storage/yourstorage/`
- [ ] Implement `core.StorageProvider` interface
- [ ] Add auto-registration in `init()` function
- [ ] Include unit tests
- [ ] Add integration tests if possible
- [ ] Update documentation (README.md, EXTENDING.md)
- [ ] Add example usage

## Coding Standards

### Go Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Use `golint` and address warnings
- Keep functions small and focused
- Write clear, descriptive variable names

### Error Handling

- Always wrap errors with context: `fmt.Errorf("operation failed: %w", err)`
- Return errors rather than panicking
- Handle errors at appropriate levels

### Comments

- Add godoc comments for all exported types and functions
- Explain "why" not "what" in comments
- Keep comments up to date with code changes

### Testing

- Write table-driven tests where appropriate
- Aim for high test coverage
- Include both unit and integration tests
- Use meaningful test names: `TestBackup_WithInvalidConfig_ReturnsError`

### Example Test Structure

```go
func TestProviderBackup(t *testing.T) {
    tests := []struct {
        name    string
        config  *core.DatabaseConfig
        wantErr bool
    }{
        {
            name: "valid configuration",
            config: &core.DatabaseConfig{
                Host: "localhost",
                // ...
            },
            wantErr: false,
        },
        {
            name: "missing host",
            config: &core.DatabaseConfig{
                // ...
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            provider, err := New(tt.config)
            if (err != nil) != tt.wantErr {
                t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Documentation

- Update README.md for new features
- Add examples to the `examples/` directory
- Update EXTENDING.md for plugin-related changes
- Include godoc comments for public APIs

## Commit Messages

Write clear, descriptive commit messages:

```text
Add support for MySQL database provider

- Implement DatabaseProvider interface for MySQL
- Add mysqldump and mysql restore commands
- Include integration tests with Docker
- Update documentation and examples

Fixes #123
```

### Commit Message Format

- Use present tense ("Add feature" not "Added feature")
- Use imperative mood ("Move cursor to..." not "Moves cursor to...")
- First line is a summary (50 chars or less)
- Optionally followed by a blank line and detailed description
- Reference issues and PRs when applicable

## Pull Request Process

1. **Update documentation** with details of changes
2. **Add tests** that prove your fix/feature works
3. **Ensure all tests pass** locally
4. **Update the README.md** if needed
5. **Follow the PR template** (we'll review and provide feedback)

### PR Title Format

- `feat: Add MongoDB database provider`
- `fix: Handle connection timeout in PostgreSQL provider`
- `docs: Update EXTENDING.md with Azure Blob Storage example`
- `test: Add integration tests for S3 provider`
- `refactor: Simplify registry lookup logic`

## Code Review Process

- All submissions require review before merging
- We aim to review PRs within 2-3 business days
- Address feedback in new commits (don't force-push during review)
- Once approved, we'll merge your PR

## Community Guidelines

- Be respectful and inclusive
- Help others learn and grow
- Give constructive feedback
- Assume good intentions

## Questions?

- **Bugs or Features**: Open an issue
- **General Questions**: Start a discussion
- **Security Issues**: Email <tuantla0708@gmail.com> (use private disclosure)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to GoArchive! 🎉
