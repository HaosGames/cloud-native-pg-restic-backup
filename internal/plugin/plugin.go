// Package plugin implements the CloudNative PostgreSQL backup plugin interface.
//
// The plugin provides HTTP endpoints for:
// - Full database backups
// - Backup restoration
// - WAL archiving
// - WAL restoration
//
// It integrates with Restic for efficient backup storage and implements
// the required interfaces for CloudNative PostgreSQL operator integration.
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

// Plugin implements the CloudNative PostgreSQL backup/restore plugin interface.
// It provides HTTP endpoints that the operator uses to manage backups and
// WAL archiving operations.
type Plugin struct {
	backupHandler  backup.Handler
	restoreHandler restore.Handler
	logger         *logging.Logger
}

// NewPlugin creates a new plugin instance with the given configuration.
// It initializes the backup and restore handlers with the provided
// Restic configuration and logger.
func NewPlugin(config restic.Config, logger *logging.Logger) *Plugin {
	client := restic.NewClient(config)
	return &Plugin{
		backupHandler:  backup.NewHandler(client),
		restoreHandler: restore.NewHandler(client),
		logger:        logger,
	}
}

// ServeHTTP implements the HTTP handler interface.
// It routes incoming requests to the appropriate handler based on the URL path:
// - /backup: Create full database backups
// - /restore: Restore from backup
// - /wal-archive: Archive WAL segments
// - /wal-restore: Restore WAL segments
func (p *Plugin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// ... [rest of the implementation remains the same]
}

// BackupRequest represents the backup API request payload.
// The operator sends this when requesting a new backup.
type BackupRequest struct {
	// BackupID is a unique identifier for this backup
	BackupID string `json:"backupID"`

	// DataFolder is the path to the PostgreSQL data directory
	DataFolder string `json:"dataFolder"`

	// DestinationPath is where the backup should be stored
	DestinationPath string `json:"destinationPath"`
}

// handleBackup processes backup requests.
// It validates the request, performs the backup operation,
// and returns the result to the operator.
func (p *Plugin) handleBackup(w http.ResponseWriter, r *http.Request, logger *logging.Logger) {
	// ... [rest of the implementation remains the same]
}

// RestoreRequest represents the restore API request payload.
// The operator sends this when requesting a backup restoration.
type RestoreRequest struct {
	// BackupID identifies the backup to restore from
	BackupID string `json:"backupID"`

	// DestFolder is where the backup should be restored
	DestFolder string `json:"destFolder"`

	// RecoveryTarget specifies PITR options if needed
	RecoveryTarget *struct {
		TargetTime     string `json:"targetTime,omitempty"`
		TargetXID      string `json:"targetXID,omitempty"`
		TargetLSN      string `json:"targetLSN,omitempty"`
		TargetName     string `json:"targetName,omitempty"`
		TargetInclusive bool   `json:"targetInclusive,omitempty"`
	} `json:"recoveryTarget,omitempty"`
}

// handleRestore processes restore requests.
// It validates the request, performs the restore operation,
// and handles any PITR requirements.
func (p *Plugin) handleRestore(w http.ResponseWriter, r *http.Request, logger *logging.Logger) {
	// ... [rest of the implementation remains the same]
}

// WALArchiveRequest represents the WAL archive API request payload.
// PostgreSQL sends this when archiving WAL segments.
type WALArchiveRequest struct {
	// WalFileName is the name of the WAL segment file
	WalFileName string `json:"walFileName"`

	// WalFilePath is the full path to the WAL segment
	WalFilePath string `json:"walFilePath"`
}

// handleWALArchive processes WAL archiving requests.
// It archives individual WAL segments using the WAL manager.
func (p *Plugin) handleWALArchive(w http.ResponseWriter, r *http.Request, logger *logging.Logger) {
	// ... [rest of the implementation remains the same]
}

// WALRestoreRequest represents the WAL restore API request payload.
// PostgreSQL sends this when requesting WAL segment restoration.
type WALRestoreRequest struct {
	// WalFileName is the name of the WAL segment to restore
	WalFileName string `json:"walFileName"`

	// DestFolder is where the WAL segment should be restored
	DestFolder string `json:"destFolder"`
}

// handleWALRestore processes WAL restoration requests.
// It restores individual WAL segments for recovery operations.
func (p *Plugin) handleWALRestore(w http.ResponseWriter, r *http.Request, logger *logging.Logger) {
	// ... [rest of the implementation remains the same]
}
