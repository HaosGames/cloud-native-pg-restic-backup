package restic

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Implementation of the Client interface using the Restic CLI

func (c *clientImpl) InitRepository(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "restic", "init")
	c.setEnvironment(cmd)

	if output, err := cmd.CombinedOutput(); err != nil {
		if strings.Contains(string(output), "repository master key and config already initialized") {
			return nil
		}
		return fmt.Errorf("failed to initialize repository: %w: %s", err, string(output))
	}
	return nil
}

func (c *clientImpl) Backup(ctx context.Context, path string, tags []string) error {
	args := []string{"backup", path}
	for _, tag := range tags {
		args = append(args, "--tag", tag)
	}

	cmd := exec.CommandContext(ctx, "restic", args...)
	c.setEnvironment(cmd)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("backup failed: %w: %s", err, string(output))
	}
	return nil
}

func (c *clientImpl) Restore(ctx context.Context, snapshotID, targetPath string) error {
	cmd := exec.CommandContext(ctx, "restic", "restore", snapshotID, "--target", targetPath)
	c.setEnvironment(cmd)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("restore failed: %w: %s", err, string(output))
	}
	return nil
}

func (c *clientImpl) RestoreFile(ctx context.Context, snapshotID, filePath, targetPath string) error {
	cmd := exec.CommandContext(ctx, "restic", "restore", snapshotID, "--include", filePath, "--target", targetPath)
	c.setEnvironment(cmd)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("file restore failed: %w: %s", err, string(output))
	}
	return nil
}

func (c *clientImpl) FindSnapshots(ctx context.Context, tags []string) ([]*Snapshot, error) {
	args := []string{"snapshots", "--json"}
	for _, tag := range tags {
		args = append(args, "--tag", tag)
	}

	cmd := exec.CommandContext(ctx, "restic", args...)
	c.setEnvironment(cmd)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %w", err)
	}

	var snapshots []*Snapshot
	if err := json.Unmarshal(output, &snapshots); err != nil {
		return nil, fmt.Errorf("failed to parse snapshots: %w", err)
	}

	return snapshots, nil
}

func (c *clientImpl) DeleteSnapshots(ctx context.Context, snapshotIDs []string) error {
	args := append([]string{"forget", "--prune"}, snapshotIDs...)
	cmd := exec.CommandContext(ctx, "restic", args...)
	c.setEnvironment(cmd)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to delete snapshots: %w: %s", err, string(output))
	}
	return nil
}

func (c *clientImpl) EnsureDirectory(ctx context.Context, path string) error {
	return os.MkdirAll(path, 0755)
}

// setEnvironment sets the required environment variables for the Restic command
func (c *clientImpl) setEnvironment(cmd *exec.Cmd) {
	cmd.Env = append(cmd.Env,
		"RESTIC_REPOSITORY="+c.config.Repository,
		"RESTIC_PASSWORD="+c.config.Password,
		"AWS_ACCESS_KEY_ID="+c.config.S3AccessKey,
		"AWS_SECRET_ACCESS_KEY="+c.config.S3SecretKey,
	)

	if c.config.S3Endpoint != "" {
		cmd.Env = append(cmd.Env, "AWS_ENDPOINT="+c.config.S3Endpoint)
	}
}
