package restic

import (
	"context"
	"fmt"
	"os/exec"
)

// Config holds the configuration for the Restic client
type Config struct {
	Repository  string
	Password    string
	S3Endpoint  string
	S3AccessKey string
	S3SecretKey string
}

// Client implements the Restic operations
type Client struct {
	config Config
}

// NewClient creates a new Restic client
func NewClient(cfg Config) *Client {
	return &Client{config: cfg}
}

// InitRepository initializes a new Restic repository if it doesn't exist
func (c *Client) InitRepository(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "restic", "init")
	c.setEnvironment(cmd)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to initialize repository: %w: %s", err, string(output))
	}
	return nil
}

// Backup creates a new backup of the specified path
func (c *Client) Backup(ctx context.Context, path string, tags []string) error {
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

// Restore restores a snapshot to the specified path
func (c *Client) Restore(ctx context.Context, snapshotID, targetPath string) error {
	cmd := exec.CommandContext(ctx, "restic", "restore", snapshotID, "--target", targetPath)
	c.setEnvironment(cmd)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("restore failed: %w: %s", err, string(output))
	}
	return nil
}

// setEnvironment sets the required environment variables for the Restic command
func (c *Client) setEnvironment(cmd *exec.Cmd) {
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
