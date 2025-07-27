package restore

import (
	"context"
	"fmt"
	"cloud-native-pg-restic-backup/internal/restic"
)

// Handler implements restore operations
type Handler struct {
	client *restic.Client
}

// NewHandler creates a new restore handler
func NewHandler(client *restic.Client) *Handler {
	return &Handler{
		client: client,
	}
}

// RestoreBackup restores a full backup to the specified directory
func (h *Handler) RestoreBackup(ctx context.Context, snapshotID, targetDir string) error {
	return h.client.Restore(ctx, snapshotID, targetDir)
}

// RestoreWAL restores a WAL segment for PITR
func (h *Handler) RestoreWAL(ctx context.Context, walFile, targetDir string) error {
	// TODO: Implement WAL restore logic
	return fmt.Errorf("WAL restore not yet implemented")
}
