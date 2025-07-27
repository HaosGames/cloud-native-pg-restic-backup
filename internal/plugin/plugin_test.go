package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"cloud-native-pg-restic-backup/internal/logging"
)

// Mock implementations
type mockBackupHandler struct {
	createBackupErr error
	archiveWALErr   error
}

func (m *mockBackupHandler) CreateBackup(_ context.Context, _ string) error {
	return m.createBackupErr
}

func (m *mockBackupHandler) ArchiveWAL(_ context.Context, _ string) error {
	return m.archiveWALErr
}

type mockRestoreHandler struct {
	restoreBackupErr error
	restoreWALErr    error
}

func (m *mockRestoreHandler) RestoreBackup(_ context.Context, _, _ string) error {
	return m.restoreBackupErr
}

func (m *mockRestoreHandler) RestoreWAL(_ context.Context, _, _ string) error {
	return m.restoreWALErr
}

// Test helper function to create a new plugin with mock handlers
func newTestPlugin() (*Plugin, *mockBackupHandler, *mockRestoreHandler) {
	backupHandler := &mockBackupHandler{}
	restoreHandler := &mockRestoreHandler{}
	logger := logging.NewLogger(logging.Config{
		Level:      "info",
		JSONOutput: false,
	})

	p := &Plugin{
		backupHandler:  backupHandler,
		restoreHandler: restoreHandler,
		logger:        logger,
	}

	return p, backupHandler, restoreHandler
}

func TestPlugin_HandleBackup(t *testing.T) {
	p, backupHandler, _ := newTestPlugin()

	tests := []struct {
		name           string
		method         string
		request        BackupRequest
		backupError    error
		expectedStatus int
	}{
		{
			name:   "successful backup",
			method: http.MethodPost,
			request: BackupRequest{
				BackupID:   "test-backup",
				DataFolder: "/data",
			},
			backupError:    nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:   "failed backup",
			method: http.MethodPost,
			request: BackupRequest{
				BackupID:   "test-backup",
				DataFolder: "/data",
			},
			backupError:    fmt.Errorf("backup failed"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "wrong method",
			method:         http.MethodGet,
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backupHandler.createBackupErr = tt.backupError

			var body []byte
			var err error
			if tt.method == http.MethodPost {
				body, err = json.Marshal(tt.request)
				if err != nil {
					t.Fatal(err)
				}
			}

			req := httptest.NewRequest(tt.method, "/backup", bytes.NewReader(body))
			w := httptest.NewRecorder()

			p.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestPlugin_HandleWALArchive(t *testing.T) {
	p, backupHandler, _ := newTestPlugin()

	tests := []struct {
		name           string
		method         string
		request        WALArchiveRequest
		archiveError   error
		expectedStatus int
	}{
		{
			name:   "successful WAL archive",
			method: http.MethodPost,
			request: WALArchiveRequest{
				WalFileName: "000000010000000000000001",
				WalFilePath: "/wal/000000010000000000000001",
			},
			archiveError:   nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:   "failed WAL archive",
			method: http.MethodPost,
			request: WALArchiveRequest{
				WalFileName: "000000010000000000000001",
				WalFilePath: "/wal/000000010000000000000001",
			},
			archiveError:   fmt.Errorf("WAL archive failed"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "wrong method",
			method:         http.MethodGet,
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backupHandler.archiveWALErr = tt.archiveError

			var body []byte
			var err error
			if tt.method == http.MethodPost {
				body, err = json.Marshal(tt.request)
				if err != nil {
					t.Fatal(err)
				}
			}

			req := httptest.NewRequest(tt.method, "/wal-archive", bytes.NewReader(body))
			w := httptest.NewRecorder()

			p.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestPlugin_HandleRestore(t *testing.T) {
	p, _, restoreHandler := newTestPlugin()

	tests := []struct {
		name           string
		method         string
		request        RestoreRequest
		restoreError   error
		expectedStatus int
	}{
		{
			name:   "successful restore",
			method: http.MethodPost,
			request: RestoreRequest{
				BackupID:   "test-backup",
				DestFolder: "/restore",
			},
			restoreError:   nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:   "failed restore",
			method: http.MethodPost,
			request: RestoreRequest{
				BackupID:   "test-backup",
				DestFolder: "/restore",
			},
			restoreError:   fmt.Errorf("restore failed"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "wrong method",
			method:         http.MethodGet,
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restoreHandler.restoreBackupErr = tt.restoreError

			var body []byte
			var err error
			if tt.method == http.MethodPost {
				body, err = json.Marshal(tt.request)
				if err != nil {
					t.Fatal(err)
				}
			}

			req := httptest.NewRequest(tt.method, "/restore", bytes.NewReader(body))
			w := httptest.NewRecorder()

			p.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
