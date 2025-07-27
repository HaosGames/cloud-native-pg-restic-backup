package restore

import (
	"context"
	"fmt"

	"cloud-native-pg-restic-backup/internal/logging"
	"cloud-native-pg-restic-backup/internal/restic"
	"cloud-native-pg-restic-backup/internal/wal"
)

// Handler interface defines the operations for restore handling
type Handler interface {
	RestoreBackup(ctx context.Context, snapshotID, targetDir string) error
	RestoreWAL(ctx context.Context, walFile, targetPath string) error
}

// handlerImpl implements the Handler interface
type handlerImpl struct {
	client     *restic.Client
	walManager *wal.Manager
	logger     *logging.Logger
}

// NewHandler creates a new restore handler
func NewHandler(client *restic.Client) Handler {
	logger := logging.NewLogger(logging.Config{
		Level:      "info",
		JSONOutput: false,
	}).Component("restore")

	return &handlerImpl{
		client:     client,
		walManager: wal.NewManager(client, logger),
		logger:     logger,
	}
}

// RestoreBackup restores a full backup to the specified directory
func (h *handlerImpl) RestoreBackup(ctx context.Context, snapshotID, targetDir string) error {
	logger := h.logger.Operation("restore_backup").WithFields(map[string]interface{}{
		"snapshot_id": snapshotID,
		"target_dir": targetDir,
	})
	logger.Info().Msg("Starting backup restore")

	if err := h.client.Restore(ctx, snapshotID, targetDir); err != nil {
		logger.Error().Err(err).Msg("Backup restore failed")
		return fmt.Errorf("failed to restore backup: %v", err)
	}

	logger.Info().Msg("Backup restore completed successfully")
	return nil
}

// RestoreWAL restores a WAL segment for PITR
func (h *handlerImpl) RestoreWAL(ctx context.Context, walFile, targetPath string) error {
	logger := h.logger.Operation("restore_wal").WithFields(map[string]interface{}{
		"wal_file": walFile,
		"target_path": targetPath,
	})
	logger.Info().Msg("Starting WAL restore")

	if err := h.walManager.RestoreWALSegment(ctx, walFile, targetPath); err != nil {
		logger.Error().Err(err).Msg("WAL restore failed")
		return fmt.Errorf("failed to restore WAL segment: %v", err)
	}

	logger.Info().Msg("WAL restore completed successfully")
	return nil
}
