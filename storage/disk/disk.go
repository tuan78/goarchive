package disk

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"goarchive/core"
)

// init registers the disk provider with the global registry
func init() {
	core.RegisterStorage("disk", func(ctx context.Context, config *core.StorageConfig) (core.StorageProvider, error) {
		return New(config)
	})
}

// Provider implements the StorageProvider interface for local disk storage
type Provider struct {
	config *core.StorageConfig
	path   string
}

// New creates a new disk storage provider
func New(config *core.StorageConfig) (*Provider, error) {
	// Use Path field, or default to ./backups
	path := config.Path
	if path == "" {
		path = "./backups"
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	return &Provider{
		config: config,
		path:   path,
	}, nil
}

// Upload saves the backup data to local disk
func (p *Provider) Upload(ctx context.Context, reader io.Reader, metadata *core.BackupMetadata) error {
	// Read all data and calculate checksum
	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read backup data: %w", err)
	}

	// Calculate MD5 checksum
	hash := md5.Sum(data)
	checksum := hex.EncodeToString(hash[:])
	metadata.Checksum = checksum
	metadata.Size = int64(len(data))

	// Create filename
	filename := p.getBackupFilename(metadata)
	fullPath := filepath.Join(p.path, filename)

	// Write to disk
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	// Write metadata file
	metadataPath := fullPath + ".meta"
	metadataContent := fmt.Sprintf(
		"ID: %s\nDatabaseName: %s\nDatabaseType: %s\nTimestamp: %s\nSize: %d\nChecksum: %s\n",
		metadata.ID,
		metadata.DatabaseName,
		metadata.DatabaseType,
		metadata.Timestamp.Format(time.RFC3339),
		metadata.Size,
		metadata.Checksum,
	)
	if err := os.WriteFile(metadataPath, []byte(metadataContent), 0644); err != nil {
		// Non-fatal, just log
		fmt.Fprintf(os.Stderr, "Warning: failed to write metadata file: %v\n", err)
	}

	return nil
}

// List lists available backups in the local directory
func (p *Provider) List(ctx context.Context) ([]*core.BackupMetadata, error) {
	entries, err := os.ReadDir(p.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []*core.BackupMetadata
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Skip metadata files
		if strings.HasSuffix(entry.Name(), ".meta") {
			continue
		}

		// Skip non-dump files
		if !strings.HasSuffix(entry.Name(), ".dump") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		backup := &core.BackupMetadata{
			ID:        entry.Name(),
			Timestamp: info.ModTime(),
			Size:      info.Size(),
		}

		// Try to read metadata file if it exists
		metadataPath := filepath.Join(p.path, entry.Name()+".meta")
		if metaData, err := os.ReadFile(metadataPath); err == nil {
			p.parseMetadata(backup, string(metaData))
		}

		backups = append(backups, backup)
	}

	// Sort by timestamp descending (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Timestamp.After(backups[j].Timestamp)
	})

	return backups, nil
}

// Download reads a backup from local disk
func (p *Provider) Download(ctx context.Context, backupID string) (io.ReadCloser, error) {
	fullPath := filepath.Join(p.path, backupID)

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("backup not found: %s", backupID)
		}
		return nil, fmt.Errorf("failed to open backup file: %w", err)
	}

	return file, nil
}

// Delete removes a backup from local disk
func (p *Provider) Delete(ctx context.Context, backupID string) error {
	fullPath := filepath.Join(p.path, backupID)

	// Delete the backup file
	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("backup not found: %s", backupID)
		}
		return fmt.Errorf("failed to delete backup file: %w", err)
	}

	// Try to delete metadata file too (non-fatal if it doesn't exist)
	metadataPath := fullPath + ".meta"
	os.Remove(metadataPath)

	return nil
}

// getBackupFilename generates the filename for a backup
func (p *Provider) getBackupFilename(metadata *core.BackupMetadata) string {
	return fmt.Sprintf("%s_%s_%s.dump",
		metadata.DatabaseName,
		metadata.DatabaseType,
		metadata.Timestamp.Format("20060102-150405"),
	)
}

// parseMetadata parses metadata from the .meta file
func (p *Provider) parseMetadata(backup *core.BackupMetadata, content string) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) != 2 {
			continue
		}
		key, value := parts[0], parts[1]

		switch key {
		case "ID":
			backup.ID = value
		case "DatabaseName":
			backup.DatabaseName = value
		case "DatabaseType":
			backup.DatabaseType = value
		case "Checksum":
			backup.Checksum = value
		case "Timestamp":
			if t, err := time.Parse(time.RFC3339, value); err == nil {
				backup.Timestamp = t
			}
		}
	}
}
