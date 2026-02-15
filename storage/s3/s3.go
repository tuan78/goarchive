package s3

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"path"
	"time"

	"goarchive/core"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// init registers the S3 provider with the global registry
func init() {
	core.RegisterStorage("s3", func(ctx context.Context, config *core.StorageConfig) (core.StorageProvider, error) {
		return New(ctx, config)
	})
}

// Provider implements the StorageProvider interface for AWS S3
type Provider struct {
	client *s3.Client
	config *core.StorageConfig
}

// New creates a new S3 provider
func New(ctx context.Context, storageConfig *core.StorageConfig) (*Provider, error) {
	// Load AWS config
	var cfg aws.Config
	var err error

	if storageConfig.AccessKey != "" && storageConfig.SecretKey != "" {
		// Use static credentials
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(storageConfig.Region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				storageConfig.AccessKey,
				storageConfig.SecretKey,
				"",
			)),
		)
	} else {
		// Use default credential chain (IAM role, env vars, etc.)
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(storageConfig.Region),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	return &Provider{
		client: client,
		config: storageConfig,
	}, nil
}

// Upload uploads the backup data to S3
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

	// Create S3 key
	key := p.getBackupKey(metadata)

	// Upload to S3
	_, err = p.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(p.config.Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/octet-stream"),
		Metadata: map[string]string{
			"database-name": metadata.DatabaseName,
			"database-type": metadata.DatabaseType,
			"backup-id":     metadata.ID,
			"timestamp":     metadata.Timestamp.Format(time.RFC3339),
			"checksum":      checksum,
		},
		Tagging: aws.String("Type=DatabaseBackup&Source=" + metadata.DatabaseType),
	})

	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	return nil
}

// List lists available backups
func (p *Provider) List(ctx context.Context) ([]*core.BackupMetadata, error) {
	result, err := p.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(p.config.Bucket),
		Prefix: aws.String(p.config.Prefix),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	var backups []*core.BackupMetadata
	for _, obj := range result.Contents {
		// Parse timestamp from key
		timestamp := *obj.LastModified

		backup := &core.BackupMetadata{
			ID:        path.Base(*obj.Key),
			Timestamp: timestamp,
			Size:      *obj.Size,
		}
		backups = append(backups, backup)
	}

	return backups, nil
}

// Download downloads a backup from S3
func (p *Provider) Download(ctx context.Context, backupID string) (io.ReadCloser, error) {
	key := path.Join(p.config.Prefix, backupID)

	result, err := p.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(p.config.Bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to download from S3: %w", err)
	}

	return result.Body, nil
}

// Delete deletes a backup from S3
func (p *Provider) Delete(ctx context.Context, backupID string) error {
	key := path.Join(p.config.Prefix, backupID)

	_, err := p.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(p.config.Bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	return nil
}

// getBackupKey generates the S3 key for a backup
func (p *Provider) getBackupKey(metadata *core.BackupMetadata) string {
	filename := fmt.Sprintf("%s_%s_%s.dump",
		metadata.DatabaseName,
		metadata.DatabaseType,
		metadata.Timestamp.Format("20060102-150405"),
	)
	return path.Join(p.config.Prefix, filename)
}
