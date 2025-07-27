package restore

import (
	"context"
	"fmt"

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
}

// NewHandler creates a new restore handler
func NewHandler(client *restic.Client) Handler {
	return &handlerImpl{
		client:     client,
		walManager: wal.NewManager(client),
	}
}

// RestoreBackup restores a full backup to the specified directory
func (h *handlerImpl) RestoreBackup(ctx context.Context, snapshotID, targetDir string) error {
	if err := h.client.Restore(ctx, snapshotID, targetDir); err != nil {
		return fmt.Errorf("failed to restore backup: %v", err)
	}
	return nil
}

// RestoreWAL restores a WAL segment for PITR
func (h *handlerImpl) RestoreWAL(ctx context.Context, walFile, targetPath string) error {
	if err := h.walManager.RestoreWALSegment(ctx, walFile, targetPath); err != nil {
		return fmt.Errorf("failed to restore WAL segment: %v", err)
	}
	return nil
}
