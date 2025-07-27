package backup

import (
	"context"
	"fmt"

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
	client     *restic.Client
	walManager *wal.Manager
}

// NewHandler creates a new backup handler
func NewHandler(client *restic.Client) Handler {
	return &handlerImpl{
		client:     client,
		walManager: wal.NewManager(client),
	}
}

// CreateBackup performs a full backup of the specified PostgreSQL data directory
func (h *handlerImpl) CreateBackup(ctx context.Context, dataDir string) error {
	// Get current WAL timeline
	timeline, err := h.walManager.GetWALTimeline(ctx)
	if err != nil {
		return fmt.Errorf("failed to get WAL timeline: %v", err)
	}

	// Create backup with timeline information
	tags := []string{
		"type:full",
		fmt.Sprintf("timeline:%d", timeline),
	}

	if err := h.client.Backup(ctx, dataDir, tags); err != nil {
		return fmt.Errorf("failed to create backup: %v", err)
	}

	return nil
}

// ArchiveWAL archives a WAL segment using Restic
func (h *handlerImpl) ArchiveWAL(ctx context.Context, walPath string) error {
	if err := h.walManager.ArchiveWAL(ctx, walPath); err != nil {
		return fmt.Errorf("failed to archive WAL: %v", err)
	}
	return nil
}
