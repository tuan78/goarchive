package postgres

import (
	"context"
	"fmt"
	"io"
	"os/exec"

	"goarchive/core"

	"github.com/jackc/pgx/v5"
)

// init registers the PostgreSQL provider with the global registry
func init() {
	core.RegisterDatabase("postgres", func(config *core.DatabaseConfig) (core.DatabaseProvider, error) {
		return New(config)
	})
}

// Provider implements the DatabaseProvider interface for PostgreSQL
type Provider struct {
	config *core.DatabaseConfig
	conn   *pgx.Conn
}

// New creates a new PostgreSQL provider
func New(config *core.DatabaseConfig) (*Provider, error) {
	connString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.Username,
		config.Password,
		config.Database,
		config.SSLMode,
	)

	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	return &Provider{
		config: config,
		conn:   conn,
	}, nil
}

// Backup creates a backup using pg_dump and returns a reader
func (p *Provider) Backup(ctx context.Context) (io.ReadCloser, error) {
	cmd := exec.CommandContext(ctx, "pg_dump",
		"-h", p.config.Host,
		"-p", fmt.Sprintf("%d", p.config.Port),
		"-U", p.config.Username,
		"-d", p.config.Database,
		"-F", "c", // Custom format
		"--no-password",
	)

	// Set PGPASSWORD environment variable
	cmd.Env = append(cmd.Env, fmt.Sprintf("PGPASSWORD=%s", p.config.Password))

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start pg_dump: %w", err)
	}

	// Return a reader that waits for the command to complete
	return &backupReader{
		ReadCloser: stdout,
		cmd:        cmd,
	}, nil
}

// Restore restores a database from backup data using pg_restore
func (p *Provider) Restore(ctx context.Context, reader io.Reader) error {
	cmd := exec.CommandContext(ctx, "pg_restore",
		"-h", p.config.Host,
		"-p", fmt.Sprintf("%d", p.config.Port),
		"-U", p.config.Username,
		"-d", p.config.Database,
		"--clean",         // Clean (drop) database objects before recreating
		"--if-exists",     // Use IF EXISTS when dropping objects
		"--no-owner",      // Skip restoration of object ownership
		"--no-privileges", // Skip restoration of access privileges
		"--no-password",
	)

	// Set PGPASSWORD environment variable
	cmd.Env = append(cmd.Env, fmt.Sprintf("PGPASSWORD=%s", p.config.Password))

	// Set stdin to the reader
	cmd.Stdin = reader

	// Execute the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restore database: %w (output: %s)", err, string(output))
	}

	return nil
}

// GetMetadata returns metadata about the database
func (p *Provider) GetMetadata() (*core.DatabaseMetadata, error) {
	var version string
	var size int64

	// Get PostgreSQL version
	err := p.conn.QueryRow(context.Background(), "SELECT version()").Scan(&version)
	if err != nil {
		return nil, fmt.Errorf("failed to get version: %w", err)
	}

	// Get database size
	err = p.conn.QueryRow(context.Background(),
		"SELECT pg_database_size($1)", p.config.Database).Scan(&size)
	if err != nil {
		return nil, fmt.Errorf("failed to get database size: %w", err)
	}

	return &core.DatabaseMetadata{
		Type:    "postgres",
		Version: version,
		Size:    size,
		Name:    p.config.Database,
	}, nil
}

// Close closes the database connection
func (p *Provider) Close() error {
	if p.conn != nil {
		return p.conn.Close(context.Background())
	}
	return nil
}

// backupReader wraps the stdout pipe and waits for the command to complete
type backupReader struct {
	io.ReadCloser
	cmd *exec.Cmd
}

// Close closes the pipe and waits for the command to finish
func (r *backupReader) Close() error {
	r.ReadCloser.Close()
	return r.cmd.Wait()
}
