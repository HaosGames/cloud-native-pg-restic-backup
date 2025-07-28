package restore

import (
	"context"
	"fmt"
	"testing"
	"time"

	"cloud-native-pg-restic-backup/internal/logging"
	"cloud-native-pg-restic-backup/internal/restic"
	"cloud-native-pg-restic-backup/internal/wal"
)

// mockResticClient implements the restic.Client interface for testing
type mockResticClient struct {
	restoreErr    error
	restoreFileErr error
	snapshots     []*restic.Snapshot
	restored      bool
	restoredFile  string
}

func (m *mockResticClient) InitRepository(_ context.Context) error {
	return nil
}

func (m *mockResticClient) Backup(_ context.Context, _ string, _ []string) error {
	return nil
}

func (m *mockResticClient) Restore(_ context.Context, _, _ string) error {
	m.restored = true
	return m.restoreErr
}

func (m *mockResticClient) RestoreFile(_ context.Context, _, file, _ string) error {
	m.restoredFile = file
	return m.restoreFileErr
}

func (m *mockResticClient) FindSnapshots(_ context.Context, _ []string) ([]*restic.Snapshot, error) {
	return m.snapshots, nil
}

func (m *mockResticClient) DeleteSnapshots(_ context.Context, _ []string) error {
	return nil
}

func (m *mockResticClient) EnsureDirectory(_ context.Context, _ string) error {
	return nil
}

func newMockResticClient() *mockResticClient {
	return &mockResticClient{
		snapshots: []*restic.Snapshot{
			{
				ID:   "test-snapshot-1",
				Time: time.Now(),
				Tags: []string{"type:wal", "000000010000000000000001"},
			},
		},
	}
}

func TestRestoreBackup(t *testing.T) {
	tests := []struct {
		name        string
		snapshotID  string
		targetDir   string
		restoreErr  error
		wantErr     bool
		wantRestore bool
	}{
		{
			name:        "successful restore",
			snapshotID:  "test-snapshot-1",
			targetDir:   "/restore",
			restoreErr:  nil,
			wantErr:     false,
			wantRestore: true,
		},
		{
			name:        "restore error",
			snapshotID:  "test-snapshot-1",
			targetDir:   "/restore",
			restoreErr:  fmt.Errorf("restore failed"),
			wantErr:     true,
			wantRestore: true,
		},
		{
			name:        "empty snapshot ID",
			snapshotID:  "",
			targetDir:   "/restore",
			restoreErr:  nil,
			wantErr:     true,
			wantRestore: false,
		},
		{
			name:        "empty target directory",
			snapshotID:  "test-snapshot-1",
			targetDir:   "",
			restoreErr:  nil,
			wantErr:     true,
			wantRestore: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := newMockResticClient()
			mockClient.restoreErr = tt.restoreErr

			// Create logger
			logger := logging.NewLogger(logging.Config{
				Level:      "info",
				JSONOutput: false,
			})

			// Create handler with mock client
			handler := &handlerImpl{
				client:     mockClient,
				walManager: wal.NewManager(mockClient, logger),
				logger:     logger,
			}

			// Execute restore
			err := handler.RestoreBackup(context.Background(), tt.snapshotID, tt.targetDir)

			// Verify results
			if (err != nil) != tt.wantErr {
				t.Errorf("RestoreBackup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify restore was called as expected
			if mockClient.restored != tt.wantRestore {
				t.Errorf("RestoreBackup() restored = %v, want %v", mockClient.restored, tt.wantRestore)
			}
		})
	}
}

func TestRestoreWAL(t *testing.T) {
	tests := []struct {
		name           string
		walFile       string
		targetPath    string
		restoreFileErr error
		wantErr       bool
	}{
		{
			name:           "successful WAL restore",
			walFile:       "000000010000000000000001",
			targetPath:    "/restore/000000010000000000000001",
			restoreFileErr: nil,
			wantErr:       false,
		},
		{
			name:           "WAL restore error",
			walFile:       "000000010000000000000001",
			targetPath:    "/restore/000000010000000000000001",
			restoreFileErr: fmt.Errorf("restore failed"),
			wantErr:       true,
		},
		{
			name:           "invalid WAL file",
			walFile:       "invalid",
			targetPath:    "/restore/invalid",
			restoreFileErr: nil,
			wantErr:       true,
		},
		{
			name:           "empty WAL file",
			walFile:       "",
			targetPath:    "/restore",
			restoreFileErr: nil,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := newMockResticClient()
			mockClient.restoreFileErr = tt.restoreFileErr

			// Create logger
			logger := logging.NewLogger(logging.Config{
				Level:      "info",
				JSONOutput: false,
			})

			// Create handler with mock client
			handler := &handlerImpl{
				client:     mockClient,
				walManager: wal.NewManager(mockClient, logger),
				logger:     logger,
			}

			// Execute WAL restore
			err := handler.RestoreWAL(context.Background(), tt.walFile, tt.targetPath)

			// Verify results
			if (err != nil) != tt.wantErr {
				t.Errorf("RestoreWAL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// For successful cases, verify correct file was restored
			if !tt.wantErr && mockClient.restoredFile != tt.walFile {
				t.Errorf("RestoreWAL() restored file = %v, want %v", mockClient.restoredFile, tt.walFile)
			}
		})
	}
}

func TestRestoreWithInvalidClient(t *testing.T) {
	// Create handler with nil client to test initialization errors
	handler := &handlerImpl{
		client: nil,
		logger: logging.NewLogger(logging.Config{
			Level:      "info",
			JSONOutput: false,
		}),
	}

	// Test restore backup with nil client
	err := handler.RestoreBackup(context.Background(), "test-snapshot", "/restore")
	if err == nil {
		t.Error("RestoreBackup() with nil client should return error")
	}

	// Test restore WAL with nil client
	err = handler.RestoreWAL(context.Background(), "000000010000000000000001", "/restore")
	if err == nil {
		t.Error("RestoreWAL() with nil client should return error")
	}
}
