package plugin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"cloud-native-pg-restic-backup/internal/backup"
	"cloud-native-pg-restic-backup/internal/logging"
	"cloud-native-pg-restic-backup/internal/restic"
	"cloud-native-pg-restic-backup/internal/restore"
)

// Plugin implements the CloudNativePG backup/restore plugin interface
type Plugin struct {
	backupHandler  backup.Handler
	restoreHandler restore.Handler
	logger         *logging.Logger
}

// NewPlugin creates a new plugin instance
func NewPlugin(config restic.Config, logger *logging.Logger) *Plugin {
	client := restic.NewClient(config)
	return &Plugin{
		backupHandler:  backup.NewHandler(client),
		restoreHandler: restore.NewHandler(client),
		logger:        logger,
	}
}

// ServeHTTP implements the HTTP handler interface
func (p *Plugin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := p.logger.Operation("http").WithFields(map[string]interface{}{
		"method": r.Method,
		"path":   r.URL.Path,
	})

	switch r.URL.Path {
	case "/backup":
		p.handleBackup(w, r, logger)
	case "/restore":
		p.handleRestore(w, r, logger)
	case "/wal-archive":
		p.handleWALArchive(w, r, logger)
	case "/wal-restore":
		p.handleWALRestore(w, r, logger)
	default:
		logger.Warn().Msg("Not found")
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// BackupRequest represents the backup API request
type BackupRequest struct {
	BackupID        string `json:"backupID"`
	DataFolder      string `json:"dataFolder"`
	DestinationPath string `json:"destinationPath"`
}

func (p *Plugin) handleBackup(w http.ResponseWriter, r *http.Request, logger *logging.Logger) {
	if r.Method != http.MethodPost {
		logger.Warn().Str("allowed_method", "POST").Msg("Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req BackupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error().Err(err).Msg("Invalid request")
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	logger = logger.WithFields(map[string]interface{}{
		"backup_id":   req.BackupID,
		"data_folder": req.DataFolder,
	})
	logger.Info().Msg("Starting backup")

	if err := p.backupHandler.CreateBackup(r.Context(), req.DataFolder); err != nil {
		logger.Error().Err(err).Msg("Backup failed")
		http.Error(w, fmt.Sprintf("Backup failed: %v", err), http.StatusInternalServerError)
		return
	}

	logger.Info().Msg("Backup completed successfully")
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

func (p *Plugin) handleRestore(w http.ResponseWriter, r *http.Request, logger *logging.Logger) {
	if r.Method != http.MethodPost {
		logger.Warn().Str("allowed_method", "POST").Msg("Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RestoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error().Err(err).Msg("Invalid request")
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	logger = logger.WithFields(map[string]interface{}{
		"backup_id":   req.BackupID,
		"dest_folder": req.DestFolder,
	})
	if req.RecoveryTarget != nil {
		logger = logger.WithFields(map[string]interface{}{
			"recovery_target": req.RecoveryTarget,
		})
	}
	logger.Info().Msg("Starting restore")

	if err := p.restoreHandler.RestoreBackup(r.Context(), req.BackupID, req.DestFolder); err != nil {
		logger.Error().Err(err).Msg("Restore failed")
		http.Error(w, fmt.Sprintf("Restore failed: %v", err), http.StatusInternalServerError)
		return
	}

	logger.Info().Msg("Restore completed successfully")
	w.WriteHeader(http.StatusOK)
}

// WALArchiveRequest represents the WAL archive API request
type WALArchiveRequest struct {
	WalFileName string `json:"walFileName"`
	WalFilePath string `json:"walFilePath"`
}

func (p *Plugin) handleWALArchive(w http.ResponseWriter, r *http.Request, logger *logging.Logger) {
	if r.Method != http.MethodPost {
		logger.Warn().Str("allowed_method", "POST").Msg("Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req WALArchiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error().Err(err).Msg("Invalid request")
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	logger = logger.WithFields(map[string]interface{}{
		"wal_file": req.WalFileName,
		"wal_path": req.WalFilePath,
	})
	logger.Info().Msg("Starting WAL archival")

	if err := p.backupHandler.ArchiveWAL(r.Context(), req.WalFilePath); err != nil {
		logger.Error().Err(err).Msg("WAL archiving failed")
		http.Error(w, fmt.Sprintf("WAL archiving failed: %v", err), http.StatusInternalServerError)
		return
	}

	logger.Info().Msg("WAL archival completed successfully")
	w.WriteHeader(http.StatusOK)
}

// WALRestoreRequest represents the WAL restore API request
type WALRestoreRequest struct {
	WalFileName string `json:"walFileName"`
	DestFolder  string `json:"destFolder"`
}

func (p *Plugin) handleWALRestore(w http.ResponseWriter, r *http.Request, logger *logging.Logger) {
	if r.Method != http.MethodPost {
		logger.Warn().Str("allowed_method", "POST").Msg("Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req WALRestoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error().Err(err).Msg("Invalid request")
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	logger = logger.WithFields(map[string]interface{}{
		"wal_file":    req.WalFileName,
		"dest_folder": req.DestFolder,
	})
	logger.Info().Msg("Starting WAL restore")

	destPath := filepath.Join(req.DestFolder, req.WalFileName)
	if err := p.restoreHandler.RestoreWAL(r.Context(), req.WalFileName, destPath); err != nil {
		logger.Error().Err(err).Msg("WAL restore failed")
		http.Error(w, fmt.Sprintf("WAL restore failed: %v", err), http.StatusInternalServerError)
		return
	}

	logger.Info().Msg("WAL restore completed successfully")
	w.WriteHeader(http.StatusOK)
}
