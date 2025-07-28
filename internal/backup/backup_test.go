package backup

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
	backupErr error
	snapshots []*restic.Snapshot
	tags      []string
}

func (m *mockResticClient) InitRepository(_ context.Context) error {
	return nil
}

func (m *mockResticClient) Backup(_ context.Context, _ string, tags []string) error {
	m.tags = tags
	return m.backupErr
}

func (m *mockResticClient) Restore(_ context.Context, _, _ string) error {
	return nil
}

func (m *mockResticClient) RestoreFile(_ context.Context, _, _, _ string) error {
	return nil
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

func TestCreateBackup(t *testing.T) {
	tests := []struct {
		name      string
		dataDir   string
		backupErr error
		wantErr   bool
	}{
		{
			name:      "successful backup",
			dataDir:   "/data",
			backupErr: nil,
			wantErr:   false,
		},
		{
			name:      "backup error",
			dataDir:   "/data",
			backupErr: fmt.Errorf("backup failed"),
			wantErr:   true,
		},
		{
			name:      "empty data directory",
			dataDir:   "",
			backupErr: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := newMockResticClient()
			mockClient.backupErr = tt.backupErr

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

			// Execute backup
			err := handler.CreateBackup(context.Background(), tt.dataDir)

			// Verify results
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateBackup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify backup was called with correct tags
			if !tt.wantErr && len(mockClient.tags) == 0 {
				t.Error("CreateBackup() did not set any tags")
			}

			// Verify backup type tag
			if !tt.wantErr {
				hasTypeTag := false
				for _, tag := range mockClient.tags {
					if tag == "type:full" {
						hasTypeTag = true
						break
					}
				}
				if !hasTypeTag {
					t.Error("CreateBackup() did not set type:full tag")
				}
			}
		})
	}
}

func TestArchiveWAL(t *testing.T) {
	tests := []struct {
		name      string
		walPath   string
		backupErr error
		wantErr   bool
	}{
		{
			name:      "successful WAL archive",
			walPath:   "/wal/000000010000000000000001",
			backupErr: nil,
			wantErr:   false,
		},
		{
			name:      "WAL archive error",
			walPath:   "/wal/000000010000000000000001",
			backupErr: fmt.Errorf("archive failed"),
			wantErr:   true,
		},
		{
			name:      "invalid WAL file name",
			walPath:   "/wal/invalid",
			backupErr: nil,
			wantErr:   true,
		},
		{
			name:      "empty WAL path",
			walPath:   "",
			backupErr: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := newMockResticClient()
			mockClient.backupErr = tt.backupErr

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

			// Execute WAL archive
			err := handler.ArchiveWAL(context.Background(), tt.walPath)

			// Verify results
			if (err != nil) != tt.wantErr {
				t.Errorf("ArchiveWAL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// For successful cases, verify WAL archiving tags
			if !tt.wantErr && len(mockClient.tags) == 0 {
				t.Error("ArchiveWAL() did not set any tags")
			}

			// Verify WAL type tag
			if !tt.wantErr {
				hasTypeTag := false
				for _, tag := range mockClient.tags {
					if tag == "type:wal" {
						hasTypeTag = true
						break
					}
				}
				if !hasTypeTag {
					t.Error("ArchiveWAL() did not set type:wal tag")
				}
			}
		})
	}
}
