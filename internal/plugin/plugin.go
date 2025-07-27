package plugin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"cloud-native-pg-restic-backup/internal/backup"
	"cloud-native-pg-restic-backup/internal/restic"
	"cloud-native-pg-restic-backup/internal/restore"
)

// Plugin implements the CloudNativePG backup/restore plugin interface
type Plugin struct {
	backupHandler  backup.Handler
	restoreHandler restore.Handler
	logger         Logger
}

// Logger interface for plugin logging
type Logger interface {
	Printf(format string, v ...interface{})
}

// NewPlugin creates a new plugin instance
func NewPlugin(config restic.Config, logger Logger) *Plugin {
	client := restic.NewClient(config)
	return &Plugin{
		backupHandler:  backup.NewHandler(client),
		restoreHandler: restore.NewHandler(client),
		logger:        logger,
	}
}

// ServeHTTP implements the HTTP handler interface
func (p *Plugin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/backup":
		p.handleBackup(w, r)
	case "/restore":
		p.handleRestore(w, r)
	case "/wal-archive":
		p.handleWALArchive(w, r)
	case "/wal-restore":
		p.handleWALRestore(w, r)
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// BackupRequest represents the backup API request
type BackupRequest struct {
	BackupID     string `json:"backupID"`
	DataFolder   string `json:"dataFolder"`
	DestinationPath string `json:"destinationPath"`
}

func (p *Plugin) handleBackup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req BackupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	p.logger.Printf("Starting backup for %s from %s", req.BackupID, req.DataFolder)

	if err := p.backupHandler.CreateBackup(r.Context(), req.DataFolder); err != nil {
		http.Error(w, fmt.Sprintf("Backup failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// RestoreRequest represents the restore API request
type RestoreRequest struct {
	BackupID    string `json:"backupID"`
	DestFolder  string `json:"destFolder"`
	RecoveryTarget *struct {
		TargetTime     string `json:"targetTime,omitempty"`
		TargetXID      string `json:"targetXID,omitempty"`
		TargetLSN      string `json:"targetLSN,omitempty"`
		TargetName     string `json:"targetName,omitempty"`
		TargetInclusive bool   `json:"targetInclusive,omitempty"`
	} `json:"recoveryTarget,omitempty"`
}

func (p *Plugin) handleRestore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RestoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	p.logger.Printf("Starting restore of backup %s to %s", req.BackupID, req.DestFolder)

	if err := p.restoreHandler.RestoreBackup(r.Context(), req.BackupID, req.DestFolder); err != nil {
		http.Error(w, fmt.Sprintf("Restore failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// WALArchiveRequest represents the WAL archive API request
type WALArchiveRequest struct {
	WalFileName string `json:"walFileName"`
	WalFilePath string `json:"walFilePath"`
}

func (p *Plugin) handleWALArchive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req WALArchiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	p.logger.Printf("Archiving WAL file %s from %s", req.WalFileName, req.WalFilePath)

	if err := p.backupHandler.ArchiveWAL(r.Context(), req.WalFilePath); err != nil {
		http.Error(w, fmt.Sprintf("WAL archiving failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// WALRestoreRequest represents the WAL restore API request
type WALRestoreRequest struct {
	WalFileName string `json:"walFileName"`
	DestFolder  string `json:"destFolder"`
}

func (p *Plugin) handleWALRestore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req WALRestoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	p.logger.Printf("Restoring WAL file %s to %s", req.WalFileName, req.DestFolder)

	destPath := filepath.Join(req.DestFolder, req.WalFileName)
	if err := p.restoreHandler.RestoreWAL(r.Context(), req.WalFileName, destPath); err != nil {
		http.Error(w, fmt.Sprintf("WAL restore failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
