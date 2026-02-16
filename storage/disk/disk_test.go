package disk_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"goarchive/core"
	"goarchive/storage/disk"
)

func TestNew(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "goarchive-new-test")
	defer os.RemoveAll(tmpDir)

	config := &core.StorageConfig{
		Type: "disk",
		Path: tmpDir,
	}

	provider, err := disk.New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if provider == nil {
		t.Error("expected non-nil provider")
	}

	// Check directory was created
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Errorf("expected directory %v to be created", tmpDir)
	}
}

func TestProvider_Upload(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "goarchive-upload-test")
	defer os.RemoveAll(tmpDir)

	config := &core.StorageConfig{
		Type: "disk",
		Path: tmpDir,
	}

	provider, err := disk.New(config)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	ctx := context.Background()
	testData := []byte("test backup data")
	reader := bytes.NewReader(testData)

	metadata := &core.BackupMetadata{
		ID:           "test-backup-20240215-120000",
		DatabaseName: "testdb",
		DatabaseType: "postgres",
		Timestamp:    time.Now(),
	}

	err = provider.Upload(ctx, reader, metadata)
	if err != nil {
		t.Errorf("Upload() error = %v", err)
	}

	// Check metadata was set
	if metadata.Checksum == "" {
		t.Error("expected checksum to be set")
	}

	if metadata.Size != int64(len(testData)) {
		t.Errorf("expected size %d, got %d", len(testData), metadata.Size)
	}
}

func TestProvider_List(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "goarchive-list-test")
	defer os.RemoveAll(tmpDir)

	config := &core.StorageConfig{
		Type: "disk",
		Path: tmpDir,
	}

	provider, err := disk.New(config)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	ctx := context.Background()

	// Initially empty
	backups, err := provider.List(ctx)
	if err != nil {
		t.Errorf("List() error = %v", err)
	}
	if len(backups) != 0 {
		t.Errorf("expected 0 backups, got %d", len(backups))
	}

	// Upload some backups
	for i := 0; i < 3; i++ {
		testData := []byte("test backup data")
		reader := bytes.NewReader(testData)

		metadata := &core.BackupMetadata{
			ID:           "test-backup",
			DatabaseName: "testdb",
			DatabaseType: "postgres",
			Timestamp:    time.Now().Add(time.Duration(i) * time.Second),
		}

		err = provider.Upload(ctx, reader, metadata)
		if err != nil {
			t.Fatalf("Upload() error = %v", err)
		}

		time.Sleep(10 * time.Millisecond)
	}

	// List backups
	backups, err = provider.List(ctx)
	if err != nil {
		t.Errorf("List() error = %v", err)
	}

	if len(backups) != 3 {
		t.Errorf("expected 3 backups, got %d", len(backups))
	}
}

func TestProvider_Download(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "goarchive-download-test")
	defer os.RemoveAll(tmpDir)

	config := &core.StorageConfig{
		Type: "disk",
		Path: tmpDir,
	}

	provider, err := disk.New(config)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	ctx := context.Background()
	testData := []byte("test backup data for download")
	reader := bytes.NewReader(testData)

	metadata := &core.BackupMetadata{
		ID:           "test-download",
		DatabaseName: "testdb",
		DatabaseType: "postgres",
		Timestamp:    time.Now(),
	}

	// Upload first
	err = provider.Upload(ctx, reader, metadata)
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}

	// List backups to get actual filename
	backups, err := provider.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(backups) == 0 {
		t.Fatal("no backups found")
	}
	// The ID returned by List is the actual filename
	backupFilename := backups[0].ID

	// Download using the filename
	downloadReader, err := provider.Download(ctx, backupFilename)
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}
	defer downloadReader.Close()

	// Verify content
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(downloadReader)
	if err != nil {
		t.Errorf("failed to read downloaded data: %v", err)
	}

	if !bytes.Equal(buf.Bytes(), testData) {
		t.Errorf("downloaded data doesn't match")
	}
}

func TestProvider_Delete(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "goarchive-delete-test")
	defer os.RemoveAll(tmpDir)

	config := &core.StorageConfig{
		Type: "disk",
		Path: tmpDir,
	}

	provider, err := disk.New(config)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	ctx := context.Background()
	testData := []byte("test backup data for deletion")
	reader := bytes.NewReader(testData)

	metadata := &core.BackupMetadata{
		ID:           "test-delete",
		DatabaseName: "testdb",
		DatabaseType: "postgres",
		Timestamp:    time.Now(),
	}

	// Upload first
	err = provider.Upload(ctx, reader, metadata)
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}

	// List backups to get actual filename
	backups, err := provider.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(backups) == 0 {
		t.Fatal("no backups found")
	}
	// The ID returned by List is the actual filename
	backupFilename := backups[0].ID

	// Delete using the filename
	err = provider.Delete(ctx, backupFilename)
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	// Verify file was deleted
	backups, _ = provider.List(ctx)
	if len(backups) != 0 {
		t.Errorf("expected 0 backups after deletion, got %d", len(backups))
	}
}

func TestProvider_AutoRegistration(t *testing.T) {
	// The disk provider should automatically register itself
	ctx := context.Background()
	config := &core.StorageConfig{
		Type: "disk",
		Path: filepath.Join(os.TempDir(), "goarchive-autoreg-test"),
	}
	defer os.RemoveAll(config.Path)

	provider, err := core.GetStorage(ctx, "disk", config)
	if err != nil {
		t.Errorf("expected disk provider to be auto-registered, got error: %v", err)
	}

	if provider == nil {
		t.Error("expected non-nil provider from auto-registration")
	}
}
