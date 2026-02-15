package core

import (
	"context"
	"io"
	"time"
)

// DatabaseProvider defines the interface for database backup operations
type DatabaseProvider interface {
	// Backup creates a backup and returns a reader for the backup data
	Backup(ctx context.Context) (io.ReadCloser, error)

	// Restore restores a database from backup data
	Restore(ctx context.Context, reader io.Reader) error

	// GetMetadata returns metadata about the database
	GetMetadata() (*DatabaseMetadata, error)

	// Close closes the database connection
	Close() error
}

// StorageProvider defines the interface for storage operations
type StorageProvider interface {
	// Upload uploads the backup data to storage
	Upload(ctx context.Context, reader io.Reader, metadata *BackupMetadata) error

	// List lists available backups
	List(ctx context.Context) ([]*BackupMetadata, error)

	// Download downloads a backup from storage
	Download(ctx context.Context, backupID string) (io.ReadCloser, error)

	// Delete deletes a backup from storage
	Delete(ctx context.Context, backupID string) error
}

// DatabaseMetadata contains information about the database
type DatabaseMetadata struct {
	Type    string
	Version string
	Size    int64
	Name    string
}

// BackupMetadata contains information about a backup
type BackupMetadata struct {
	ID           string
	DatabaseName string
	DatabaseType string
	Timestamp    time.Time
	Size         int64
	Checksum     string
	Tags         map[string]string
}

// BackupService orchestrates the backup process
type BackupService struct {
	database DatabaseProvider
	storage  StorageProvider
}

// NewBackupService creates a new backup service
func NewBackupService(db DatabaseProvider, storage StorageProvider) *BackupService {
	return &BackupService{
		database: db,
		storage:  storage,
	}
}

// Execute performs the backup operation
func (s *BackupService) Execute(ctx context.Context) (*BackupMetadata, error) {
	// Get database metadata
	dbMeta, err := s.database.GetMetadata()
	if err != nil {
		return nil, err
	}

	// Create backup
	reader, err := s.database.Backup(ctx)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// Prepare backup metadata
	metadata := &BackupMetadata{
		ID:           generateBackupID(),
		DatabaseName: dbMeta.Name,
		DatabaseType: dbMeta.Type,
		Timestamp:    time.Now(),
		Tags:         make(map[string]string),
	}

	// Upload to storage
	if err := s.storage.Upload(ctx, reader, metadata); err != nil {
		return nil, err
	}

	return metadata, nil
}

// Restore performs the restore operation
func (s *BackupService) Restore(ctx context.Context, backupID string) error {
	// Download from storage
	reader, err := s.storage.Download(ctx, backupID)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Restore to database
	if err := s.database.Restore(ctx, reader); err != nil {
		return err
	}

	return nil
}

// List lists all available backups
func (s *BackupService) List(ctx context.Context) ([]*BackupMetadata, error) {
	return s.storage.List(ctx)
}

// Delete deletes a backup
func (s *BackupService) Delete(ctx context.Context, backupID string) error {
	return s.storage.Delete(ctx, backupID)
}

// generateBackupID generates a unique backup identifier
func generateBackupID() string {
	return time.Now().Format("20060102-150405")
}
