// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"time"

	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/volume"
	"github.com/daytonaio/runner/pkg/volume/multiplexer"
)

const (
	multiplexerMountPath = "/mnt/daytona-volumes"
	multiplexerAddress   = "unix:///var/run/daytona-volume-multiplexer.sock"
)

// VolumeMultiplexerConfig holds configuration for the volume multiplexer
type VolumeMultiplexerConfig struct {
	Enabled    bool
	Address    string
	MountPath  string
	CacheDir   string
	MaxCacheGB int
}

// ensureMultiplexerRunning ensures the multiplexer daemon is running
func (d *DockerClient) ensureMultiplexerRunning(ctx context.Context) error {
	// Check if multiplexer is already running
	client, err := multiplexer.NewClient(multiplexerAddress)
	if err == nil {
		err = client.HealthCheck(ctx)
		client.Close()
		if err == nil {
			d.logger.Debug("Volume multiplexer already running")
			return nil
		}
	}

	// Start multiplexer daemon
	d.logger.Info("Starting volume multiplexer daemon")

	// Create directories
	os.MkdirAll(multiplexerMountPath, 0755)
	os.MkdirAll("/var/run", 0755)
	os.MkdirAll("/var/cache/daytona-volumes", 0755)

	// Start the daemon process
	cmd := exec.Command(
		"daytona-volume-multiplexer",
		"--mount-path", multiplexerMountPath,
		"--grpc-address", multiplexerAddress,
		"--cache-dir", "/var/cache/daytona-volumes",
		"--max-cache-gb", "10",
	)

	// Use systemd-run if available to isolate the daemon
	if _, err := os.Stat("/run/systemd/system"); err == nil {
		sdArgs := []string{
			"--scope",
			"--property=Restart=on-failure",
			"--property=RestartSec=5s",
			"--",
			"daytona-volume-multiplexer",
			"--mount-path", multiplexerMountPath,
			"--grpc-address", multiplexerAddress,
			"--cache-dir", "/var/cache/daytona-volumes",
			"--max-cache-gb", "10",
		}
		cmd = exec.Command("systemd-run", sdArgs...)
	}

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start multiplexer daemon: %w", err)
	}

	// Wait for daemon to be ready
	maxAttempts := 30
	for i := 0; i < maxAttempts; i++ {
		client, err := multiplexer.NewClient(multiplexerAddress)
		if err == nil {
			err = client.HealthCheck(ctx)
			client.Close()
			if err == nil {
				d.logger.Info("Volume multiplexer daemon started successfully")
				return nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("multiplexer daemon did not become ready within timeout")
}

// getVolumesMountPathBindsMultiplexer handles volume mounts using the multiplexer
func (d *DockerClient) getVolumesMountPathBindsMultiplexer(ctx context.Context, volumes []dto.VolumeDTO) ([]string, error) {
	// Ensure multiplexer is running
	if err := d.ensureMultiplexerRunning(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure multiplexer running: %w", err)
	}

	// Connect to multiplexer
	client, err := multiplexer.NewClient(multiplexerAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to multiplexer: %w", err)
	}
	defer client.Close()

	volumeMountPathBinds := make([]string, 0)

	// Register volumes with multiplexer
	for _, vol := range volumes {
		// Create provider config
		config := volume.ProviderConfig{
			Type:       "s3",
			Endpoint:   d.awsEndpointUrl,
			AccessKey:  d.awsAccessKeyId,
			SecretKey:  d.awsSecretAccessKey,
			Region:     d.awsRegion,
			BucketName: vol.VolumeId,
		}

		// Handle subpath
		if vol.Subpath != nil && *vol.Subpath != "" {
			config.Subpath = *vol.Subpath
		}

		// Register volume
		err = client.RegisterVolume(ctx, vol.VolumeId, config, false)
		if err != nil {
			return nil, fmt.Errorf("failed to register volume %s: %w", vol.VolumeId, err)
		}

		// Increment reference count
		err = client.IncrementRefCount(ctx, vol.VolumeId)
		if err != nil {
			return nil, fmt.Errorf("failed to increment ref count for volume %s: %w", vol.VolumeId, err)
		}

		// Build bind mount path
		sourcePath := filepath.Join(multiplexerMountPath, vol.VolumeId)
		if vol.Subpath != nil && *vol.Subpath != "" {
			sourcePath = filepath.Join(sourcePath, *vol.Subpath)
		}

		// Ensure source directory exists (multiplexer creates it virtually)
		// Wait a bit for FUSE to populate the directory
		time.Sleep(50 * time.Millisecond)

		d.logger.DebugContext(ctx, "binding multiplexed volume",
			"volumeId", vol.VolumeId,
			"sourcePath", sourcePath,
			"mountPath", vol.MountPath)

		volumeMountPathBinds = append(volumeMountPathBinds,
			fmt.Sprintf("%s/:%s/", sourcePath, vol.MountPath))
	}

	return volumeMountPathBinds, nil
}

// releaseMultiplexedVolumes decrements reference counts for volumes
func (d *DockerClient) releaseMultiplexedVolumes(ctx context.Context, volumeIds []string) {
	client, err := multiplexer.NewClient(multiplexerAddress)
	if err != nil {
		d.logger.Error("Failed to connect to multiplexer for volume release", "error", err)
		return
	}
	defer client.Close()

	for _, volumeId := range volumeIds {
		err = client.DecrementRefCount(ctx, volumeId)
		if err != nil {
			d.logger.Error("Failed to decrement ref count", "volumeId", volumeId, "error", err)
		}
	}
}

// Add feature flag check to DockerClient methods
func (d *DockerClient) useVolumeMultiplexer() bool {
	// Check environment variable or config
	return os.Getenv("USE_VOLUME_MULTIPLEXER") == "true"
}

// Modified getVolumesMountPathBinds to use multiplexer when enabled
func (d *DockerClient) getVolumesMountPathBindsWithMultiplexer(ctx context.Context, volumes []dto.VolumeDTO) ([]string, error) {
	if d.useVolumeMultiplexer() {
		return d.getVolumesMountPathBindsMultiplexer(ctx, volumes)
	}
	// Fall back to original implementation
	return d.getVolumesMountPathBinds(ctx, volumes)
}
