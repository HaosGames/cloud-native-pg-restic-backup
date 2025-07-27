package restore

import (
	"context"
	"fmt"
	"cloud-native-pg-restic-backup/internal/restic"
)

// Handler interface defines the operations for restore handling
type Handler interface {
	RestoreBackup(ctx context.Context, snapshotID, targetDir string) error
	RestoreWAL(ctx context.Context, walFile, targetPath string) error
}

// handlerImpl implements the Handler interface
type handlerImpl struct {
	client *restic.Client
}

// NewHandler creates a new restore handler
func NewHandler(client *restic.Client) Handler {
	return &handlerImpl{
		client: client,
	}
}

// RestoreBackup restores a full backup to the specified directory
func (h *handlerImpl) RestoreBackup(ctx context.Context, snapshotID, targetDir string) error {
	return h.client.Restore(ctx, snapshotID, targetDir)
}

// RestoreWAL restores a WAL segment for PITR
func (h *handlerImpl) RestoreWAL(ctx context.Context, walFile, targetPath string) error {
	// TODO: Implement WAL restore logic
	// This will need to:
	// 1. Find the snapshot containing the WAL file
	// 2. Restore only that specific WAL file
	// 3. Place it in the correct location for PostgreSQL recovery
	return fmt.Errorf("WAL restore not yet implemented")
}
