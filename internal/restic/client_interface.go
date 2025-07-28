package restic

import (
	"context"
	"time"
)

// Client defines the interface for Restic operations
type Client interface {
	// InitRepository initializes a new Restic repository
	InitRepository(ctx context.Context) error

	// Backup creates a new backup of the specified path
	Backup(ctx context.Context, path string, tags []string) error

	// Restore restores a snapshot to the specified path
	Restore(ctx context.Context, snapshotID, targetPath string) error

	// RestoreFile restores a single file from a snapshot
	RestoreFile(ctx context.Context, snapshotID, filePath, targetPath string) error

	// FindSnapshots finds snapshots matching the given tags
	FindSnapshots(ctx context.Context, tags []string) ([]*Snapshot, error)

	// DeleteSnapshots deletes the specified snapshots
	DeleteSnapshots(ctx context.Context, snapshotIDs []string) error

	// EnsureDirectory ensures a directory exists
	EnsureDirectory(ctx context.Context, path string) error
}

// Snapshot represents a Restic snapshot
type Snapshot struct {
	ID       string    `json:"id"`
	Time     time.Time `json:"time"`
	Hostname string    `json:"hostname"`
	Tags     []string  `json:"tags"`
}

// Config holds the configuration for the Restic client
type Config struct {
	Repository  string
	Password    string
	S3Endpoint  string
	S3AccessKey string
	S3SecretKey string
}

// clientImpl implements the Client interface using the Restic CLI
type clientImpl struct {
	config Config
}

// NewClient creates a new Restic client
func NewClient(cfg Config) Client {
	return &clientImpl{
		config: cfg,
	}
}
