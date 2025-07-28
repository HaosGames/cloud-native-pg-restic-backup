package tests

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"cloud-native-pg-restic-backup/internal/backup"
	"cloud-native-pg-restic-backup/internal/restic"
)

func TestBackupRestore(t *testing.T) {
	// Skip if not in CI/CD or explicit test environment
	if os.Getenv("TEST_RESTIC_REPOSITORY") == "" {
		t.Skip("Skipping integration test - TEST_RESTIC_REPOSITORY not set")
	}

	ctx := context.Background()

	// Create test configuration
	config := restic.Config{
		Repository:  os.Getenv("TEST_RESTIC_REPOSITORY"),
		Password:    os.Getenv("TEST_RESTIC_PASSWORD"),
		S3Endpoint:  os.Getenv("TEST_S3_ENDPOINT"),
		S3AccessKey: os.Getenv("TEST_AWS_ACCESS_KEY_ID"),
		S3SecretKey: os.Getenv("TEST_AWS_SECRET_ACCESS_KEY"),
	}

	// Create test data directory
	testDataDir := filepath.Join(t.TempDir(), "data")
	if err := os.MkdirAll(testDataDir, 0755); err != nil {
		t.Fatalf("Failed to create test data directory: %v", err)
	}

	// Create some test files
	testFile := filepath.Join(testDataDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test data"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Initialize client and handlers
	client := restic.NewClient(config)
	backupHandler := backup.NewHandler(client)

	// Test backup
	t.Run("Backup", func(t *testing.T) {
		err := backupHandler.CreateBackup(ctx, testDataDir)
		if err != nil {
			t.Fatalf("Backup failed: %v", err)
		}
	})

	// Test restore
	t.Run("Restore", func(t *testing.T) {
		restoreDir := filepath.Join(t.TempDir(), "restore")
		if err := os.MkdirAll(restoreDir, 0755); err != nil {
			t.Fatalf("Failed to create restore directory: %v", err)
		}

		// TODO: Implement restore test once restore functionality is complete
		t.Skip("Restore test not implemented yet")
	})
}

func TestWALArchiving(t *testing.T) {
	if os.Getenv("TEST_RESTIC_REPOSITORY") == "" {
		t.Skip("Skipping integration test - TEST_RESTIC_REPOSITORY not set")
	}

	ctx := context.Background()

	// Create test configuration
	config := restic.Config{
		Repository:  os.Getenv("TEST_RESTIC_REPOSITORY"),
		Password:    os.Getenv("TEST_RESTIC_PASSWORD"),
		S3Endpoint:  os.Getenv("TEST_S3_ENDPOINT"),
		S3AccessKey: os.Getenv("TEST_AWS_ACCESS_KEY_ID"),
		S3SecretKey: os.Getenv("TEST_AWS_SECRET_ACCESS_KEY"),
	}

	// Initialize client and handlers
	client := restic.NewClient(config)
	backupHandler := backup.NewHandler(client)

	// Create test WAL file
	testWALDir := filepath.Join(t.TempDir(), "wal")
	if err := os.MkdirAll(testWALDir, 0755); err != nil {
		t.Fatalf("Failed to create test WAL directory: %v", err)
	}

	testWALFile := filepath.Join(testWALDir, "000000010000000000000001")
	if err := os.WriteFile(testWALFile, []byte("test WAL data"), 0644); err != nil {
		t.Fatalf("Failed to create test WAL file: %v", err)
	}

	// Test WAL archiving
	t.Run("ArchiveWAL", func(t *testing.T) {
		err := backupHandler.ArchiveWAL(ctx, testWALFile)
		if err != nil {
			t.Fatalf("WAL archiving failed: %v", err)
		}
	})
}
