package backup

import (
	"context"
	"cloud-native-pg-restic-backup/internal/restic"
)

// Handler interface defines the operations for backup handling
type Handler interface {
	CreateBackup(ctx context.Context, dataDir string) error
	ArchiveWAL(ctx context.Context, walPath string) error
}

// handlerImpl implements the Handler interface
type handlerImpl struct {
	client *restic.Client
}

// NewHandler creates a new backup handler
func NewHandler(client *restic.Client) Handler {
	return &handlerImpl{
		client: client,
	}
}

// CreateBackup performs a full backup of the specified PostgreSQL data directory
func (h *handlerImpl) CreateBackup(ctx context.Context, dataDir string) error {
	return h.client.Backup(ctx, dataDir, []string{"type:full"})
}

// ArchiveWAL archives a WAL segment using Restic
func (h *handlerImpl) ArchiveWAL(ctx context.Context, walPath string) error {
	return h.client.Backup(ctx, walPath, []string{"type:wal"})
}
