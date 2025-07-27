package backup

import (
	"context"
	"cloud-native-pg-restic-backup/internal/restic"
)

// Handler implements backup operations
type Handler struct {
	client *restic.Client
}

// NewHandler creates a new backup handler
func NewHandler(client *restic.Client) *Handler {
	return &Handler{
		client: client,
	}
}

// CreateBackup performs a full backup of the specified PostgreSQL data directory
func (h *Handler) CreateBackup(ctx context.Context, dataDir string) error {
	return h.client.Backup(ctx, dataDir, []string{"type:full"})
}

// ArchiveWAL archives a WAL segment using Restic
func (h *Handler) ArchiveWAL(ctx context.Context, walPath string) error {
	return h.client.Backup(ctx, walPath, []string{"type:wal"})
}
