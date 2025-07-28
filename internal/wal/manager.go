// Package wal provides WAL (Write-Ahead Log) management functionality for PostgreSQL.
package wal

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"cloud-native-pg-restic-backup/internal/logging"
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
	logger *logging.Logger
}

// NewManager creates a new WAL manager
func NewManager(client *restic.Client, logger *logging.Logger) *Manager {
	return &Manager{
		client: client,
		logger: logger.Component("wal"),
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
	logger := m.logger.Operation("archive_wal").WithFields(map[string]interface{}{
		"wal_path": walPath,
	})

	walFileName := filepath.Base(walPath)
	segment, err := ParseWALFileName(walFileName)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to parse WAL file name")
		return fmt.Errorf("failed to parse WAL file name: %v", err)
	}

	logger = logger.WithFields(map[string]interface{}{
		"timeline":    segment.Timeline,
		"logical_id":  segment.LogicalID,
		"segment_id":  segment.SegmentID,
		"wal_file":    walFileName,
	})

	logger.Info().Msg("Starting WAL segment archival")

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
		logger.Error().Err(err).Msg("Failed to archive WAL segment")
		return fmt.Errorf("failed to archive WAL segment: %v", err)
	}

	logger.Info().Msg("Successfully archived WAL segment")
	return nil
}

// FindWALSegment finds a specific WAL segment in the repository
func (m *Manager) FindWALSegment(ctx context.Context, walFileName string) (*Segment, error) {
	logger := m.logger.Operation("find_wal").WithFields(map[string]interface{}{
		"wal_file": walFileName,
	})

	logger.Info().Msg("Searching for WAL segment")

	segment, err := ParseWALFileName(walFileName)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to parse WAL file name")
		return nil, fmt.Errorf("failed to parse WAL file name: %v", err)
	}

	// Find snapshots with matching WAL file tag
	snapshots, err := m.client.FindSnapshots(ctx, []string{
		"type:wal",
		fmt.Sprintf("wal_file:%s", walFileName),
	})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to find WAL segment")
		return nil, fmt.Errorf("failed to find WAL segment: %v", err)
	}

	if len(snapshots) == 0 {
		logger.Error().Msg("WAL segment not found")
		return nil, fmt.Errorf("WAL segment not found: %s", walFileName)
	}

	// Use the most recent snapshot if multiple exist
	latestSnapshot := snapshots[0]
	segment.BackupID = latestSnapshot.ID
	segment.ArchivedAt = latestSnapshot.Time

	logger.Info().
		Str("backup_id", segment.BackupID).
		Time("archived_at", segment.ArchivedAt).
		Msg("Found WAL segment")

	return segment, nil
}

// RestoreWALSegment restores a specific WAL segment
func (m *Manager) RestoreWALSegment(ctx context.Context, walFileName, targetPath string) error {
	logger := m.logger.Operation("restore_wal").WithFields(map[string]interface{}{
		"wal_file": walFileName,
		"target_path": targetPath,
	})

	logger.Info().Msg("Starting WAL segment restoration")

	segment, err := m.FindWALSegment(ctx, walFileName)
	if err != nil {
		return err
	}

	// Ensure target directory exists
	targetDir := filepath.Dir(targetPath)
	if err := m.client.EnsureDirectory(ctx, targetDir); err != nil {
		logger.Error().Err(err).Msg("Failed to create target directory")
		return fmt.Errorf("failed to create target directory: %v", err)
	}

	// Restore only the specific WAL file
	if err := m.client.RestoreFile(ctx, segment.BackupID, walFileName, targetPath); err != nil {
		logger.Error().Err(err).Msg("Failed to restore WAL segment")
		return fmt.Errorf("failed to restore WAL segment: %v", err)
	}

	logger.Info().Msg("Successfully restored WAL segment")
	return nil
}

// CleanupWALSegments removes WAL segments before a given time
func (m *Manager) CleanupWALSegments(ctx context.Context, before time.Time) error {
	logger := m.logger.Operation("cleanup_wal").WithFields(map[string]interface{}{
		"before": before,
	})

	logger.Info().Msg("Starting WAL segments cleanup")

	// Find all WAL snapshots before the specified time
	snapshots, err := m.client.FindSnapshots(ctx, []string{"type:wal"})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to list WAL segments")
		return fmt.Errorf("failed to list WAL segments: %v", err)
	}

	var snapshotsToDelete []string
	for _, snapshot := range snapshots {
		if snapshot.Time.Before(before) {
			snapshotsToDelete = append(snapshotsToDelete, snapshot.ID)
		}
	}

	logger.Info().
		Int("total_segments", len(snapshots)).
		Int("segments_to_delete", len(snapshotsToDelete)).
		Msg("Found WAL segments for cleanup")

	if len(snapshotsToDelete) > 0 {
		if err := m.client.DeleteSnapshots(ctx, snapshotsToDelete); err != nil {
			logger.Error().Err(err).Msg("Failed to delete old WAL segments")
			return fmt.Errorf("failed to delete old WAL segments: %v", err)
		}
		logger.Info().Msg("Successfully deleted old WAL segments")
	} else {
		logger.Info().Msg("No WAL segments to delete")
	}

	return nil
}

// GetWALTimeline returns the current WAL timeline
func (m *Manager) GetWALTimeline(ctx context.Context) (Timeline, error) {
	logger := m.logger.Operation("get_timeline")
	logger.Info().Msg("Getting current WAL timeline")

	// Find the most recent WAL segment
	snapshots, err := m.client.FindSnapshots(ctx, []string{"type:wal"})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get WAL timeline")
		return 0, fmt.Errorf("failed to get WAL timeline: %v", err)
	}

	if len(snapshots) == 0 {
		logger.Info().Msg("No WAL segments found, using default timeline 1")
		return 1, nil // Default timeline if no WAL segments exist
	}

	// Parse the WAL file name from the most recent snapshot
	for _, tag := range snapshots[0].Tags {
		if walFile := walFileRegex.FindString(tag); walFile != "" {
			segment, err := ParseWALFileName(walFile)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to parse WAL file name")
				return 0, fmt.Errorf("failed to parse WAL file name: %v", err)
			}
			logger.Info().
				Uint32("timeline", uint32(segment.Timeline)).
				Msg("Found current WAL timeline")
			return segment.Timeline, nil
		}
	}

	logger.Error().Msg("No valid WAL file found in latest snapshot")
	return 0, fmt.Errorf("no valid WAL file found in latest snapshot")
}
