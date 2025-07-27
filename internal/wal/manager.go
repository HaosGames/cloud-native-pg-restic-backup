package wal

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"cloud-native-pg-restic-backup/internal/restic"
)

// Timeline represents a PostgreSQL WAL timeline
type Timeline uint32

// LSN represents a PostgreSQL Log Sequence Number
type LSN uint64

// Segment represents a WAL segment file
type Segment struct {
	Timeline    Timeline
	LogicalID   uint64
	SegmentID   uint64
	Path        string
	BackupID    string
	ArchivedAt  time.Time
}

var (
	// Example WAL file name: 000000010000000000000001
	walFileRegex = regexp.MustCompile(`^([0-9A-F]{8})([0-9A-F]{8})([0-9A-F]{8})$`)
)

// Manager handles WAL segment operations
type Manager struct {
	client *restic.Client
}

// NewManager creates a new WAL manager
func NewManager(client *restic.Client) *Manager {
	return &Manager{
		client: client,
	}
}

// ParseWALFileName parses a WAL file name into its components
func ParseWALFileName(name string) (*Segment, error) {
	matches := walFileRegex.FindStringSubmatch(name)
	if matches == nil {
		return nil, fmt.Errorf("invalid WAL file name format: %s", name)
	}

	timeline, err := strconv.ParseUint(matches[1], 16, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid timeline: %v", err)
	}

	logicalID, err := strconv.ParseUint(matches[2], 16, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid logical ID: %v", err)
	}

	segmentID, err := strconv.ParseUint(matches[3], 16, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid segment ID: %v", err)
	}

	return &Segment{
		Timeline:  Timeline(timeline),
		LogicalID: logicalID,
		SegmentID: segmentID,
	}, nil
}

// ArchiveWAL archives a WAL segment
func (m *Manager) ArchiveWAL(ctx context.Context, walPath string) error {
	walFileName := filepath.Base(walPath)
	segment, err := ParseWALFileName(walFileName)
	if err != nil {
		return fmt.Errorf("failed to parse WAL file name: %v", err)
	}

	// Set tags for WAL segment identification
	tags := []string{
		"type:wal",
		fmt.Sprintf("timeline:%d", segment.Timeline),
		fmt.Sprintf("logical_id:%d", segment.LogicalID),
		fmt.Sprintf("segment_id:%d", segment.SegmentID),
		fmt.Sprintf("wal_file:%s", walFileName),
	}

	// Archive the WAL segment
	if err := m.client.Backup(ctx, walPath, tags); err != nil {
		return fmt.Errorf("failed to archive WAL segment: %v", err)
	}

	return nil
}

// FindWALSegment finds a specific WAL segment in the repository
func (m *Manager) FindWALSegment(ctx context.Context, walFileName string) (*Segment, error) {
	segment, err := ParseWALFileName(walFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse WAL file name: %v", err)
	}

	// Find snapshots with matching WAL file tag
	snapshots, err := m.client.FindSnapshots(ctx, []string{
		"type:wal",
		fmt.Sprintf("wal_file:%s", walFileName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find WAL segment: %v", err)
	}

	if len(snapshots) == 0 {
		return nil, fmt.Errorf("WAL segment not found: %s", walFileName)
	}

	// Use the most recent snapshot if multiple exist
	latestSnapshot := snapshots[0]
	segment.BackupID = latestSnapshot.ID
	segment.ArchivedAt = latestSnapshot.Time

	return segment, nil
}

// RestoreWALSegment restores a specific WAL segment
func (m *Manager) RestoreWALSegment(ctx context.Context, walFileName, targetPath string) error {
	segment, err := m.FindWALSegment(ctx, walFileName)
	if err != nil {
		return err
	}

	// Ensure target directory exists
	targetDir := filepath.Dir(targetPath)
	if err := m.client.EnsureDirectory(ctx, targetDir); err != nil {
		return fmt.Errorf("failed to create target directory: %v", err)
	}

	// Restore only the specific WAL file
	if err := m.client.RestoreFile(ctx, segment.BackupID, walFileName, targetPath); err != nil {
		return fmt.Errorf("failed to restore WAL segment: %v", err)
	}

	return nil
}

// CleanupWALSegments removes WAL segments before a given time
func (m *Manager) CleanupWALSegments(ctx context.Context, before time.Time) error {
	// Find all WAL snapshots before the specified time
	snapshots, err := m.client.FindSnapshots(ctx, []string{"type:wal"})
	if err != nil {
		return fmt.Errorf("failed to list WAL segments: %v", err)
	}

	var snapshotsToDelete []string
	for _, snapshot := range snapshots {
		if snapshot.Time.Before(before) {
			snapshotsToDelete = append(snapshotsToDelete, snapshot.ID)
		}
	}

	if len(snapshotsToDelete) > 0 {
		if err := m.client.DeleteSnapshots(ctx, snapshotsToDelete); err != nil {
			return fmt.Errorf("failed to delete old WAL segments: %v", err)
		}
	}

	return nil
}

// GetWALTimeline returns the current WAL timeline
func (m *Manager) GetWALTimeline(ctx context.Context) (Timeline, error) {
	// Find the most recent WAL segment
	snapshots, err := m.client.FindSnapshots(ctx, []string{"type:wal"})
	if err != nil {
		return 0, fmt.Errorf("failed to get WAL timeline: %v", err)
	}

	if len(snapshots) == 0 {
		return 1, nil // Default timeline if no WAL segments exist
	}

	// Parse the WAL file name from the most recent snapshot
	for _, tag := range snapshots[0].Tags {
		if walFile := walFileRegex.FindString(tag); walFile != "" {
			segment, err := ParseWALFileName(walFile)
			if err != nil {
				return 0, fmt.Errorf("failed to parse WAL file name: %v", err)
			}
			return segment.Timeline, nil
		}
	}

	return 0, fmt.Errorf("no valid WAL file found in latest snapshot")
}
