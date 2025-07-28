package backup

import (
	"context"
	"fmt"

	"cloud-native-pg-restic-backup/internal/logging"
	"cloud-native-pg-restic-backup/internal/restic"
	"cloud-native-pg-restic-backup/internal/wal"
)

// Handler interface defines the operations for backup handling
type Handler interface {
	CreateBackup(ctx context.Context, dataDir string) error
	ArchiveWAL(ctx context.Context, walPath string) error
}

// handlerImpl implements the Handler interface
type handlerImpl struct {
	client     restic.Client
	walManager *wal.Manager
	logger     *logging.Logger
}

// NewHandler creates a new backup handler
func NewHandler(client restic.Client) Handler {
	logger := logging.NewLogger(logging.Config{
		Level:      "info",
		JSONOutput: false,
	}).Component("backup")

	return &handlerImpl{
		client:     client,
		walManager: wal.NewManager(client, logger),
		logger:     logger,
	}
}

// CreateBackup performs a full backup of the specified PostgreSQL data directory
func (h *handlerImpl) CreateBackup(ctx context.Context, dataDir string) error {
	if dataDir == "" {
		return fmt.Errorf("data directory not specified")
	}

	logger := h.logger.Operation("create_backup").WithFields(map[string]interface{}{
		"data_dir": dataDir,
	})

	// Get current WAL timeline
	timeline, err := h.walManager.GetWALTimeline(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get WAL timeline")
		return fmt.Errorf("failed to get WAL timeline: %v", err)
	}

	logger = logger.WithFields(map[string]interface{}{
		"timeline": timeline,
	})
	logger.Info().Msg("Starting backup")

	// Create backup with timeline information
	tags := []string{
		"type:full",
		fmt.Sprintf("timeline:%d", timeline),
	}

	if err := h.client.Backup(ctx, dataDir, tags); err != nil {
		logger.Error().Err(err).Msg("Backup failed")
		return fmt.Errorf("failed to create backup: %v", err)
	}

	logger.Info().Msg("Backup completed successfully")
	return nil
}

// ArchiveWAL archives a WAL segment using Restic
func (h *handlerImpl) ArchiveWAL(ctx context.Context, walPath string) error {
	if walPath == "" {
		return fmt.Errorf("WAL path not specified")
	}

	logger := h.logger.Operation("archive_wal").WithFields(map[string]interface{}{
		"wal_path": walPath,
	})

	logger.Info().Msg("Starting WAL archival")

	if err := h.walManager.ArchiveWAL(ctx, walPath); err != nil {
		logger.Error().Err(err).Msg("WAL archival failed")
		return fmt.Errorf("failed to archive WAL: %v", err)
	}

	logger.Info().Msg("WAL archival completed successfully")
	return nil
}
