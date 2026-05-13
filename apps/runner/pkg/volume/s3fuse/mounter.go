// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package s3fuse

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"time"

	"github.com/daytonaio/common-go/pkg/log"
)

// Config holds the AWS credentials and endpoint needed to mount S3 buckets.
type Config struct {
	AWSRegion          string
	AWSEndpointUrl     string
	AWSAccessKeyId     string
	AWSSecretAccessKey string
}

// Mounter implements volume.Mounter using mount-s3 (AWS Mountpoint for S3).
type Mounter struct {
	cfg    Config
	logger *slog.Logger
}

func NewMounter(cfg Config, logger *slog.Logger) *Mounter {
	return &Mounter{
		cfg:    cfg,
		logger: logger.With(slog.String("component", "s3fuse-mounter")),
	}
}

func (m *Mounter) Mount(ctx context.Context, volumeID string, mountPath string) error {
	if m.IsMounted(mountPath) {
		m.logger.DebugContext(ctx, "volume already mounted", "volumeId", volumeID, "mountPath", mountPath)
		return nil
	}

	m.logger.InfoContext(ctx, "mounting S3 volume", "volumeId", volumeID, "mountPath", mountPath)

	cmd := m.buildMountCmd(ctx, volumeID, mountPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mount-s3 failed for %s at %s: %w", volumeID, mountPath, err)
	}

	return nil
}

func (m *Mounter) Unmount(_ context.Context, mountPath string) error {
	return exec.Command("umount", mountPath).Run()
}

func (m *Mounter) IsMounted(mountPath string) bool {
	return exec.Command("mountpoint", mountPath).Run() == nil
}

func (m *Mounter) WaitUntilReady(ctx context.Context, mountPath string) error {
	const maxAttempts = 50
	const sleepDuration = 100 * time.Millisecond

	for i := range maxAttempts {
		if !m.IsMounted(mountPath) {
			return fmt.Errorf("mount disappeared during readiness check")
		}

		if _, err := os.Stat(mountPath); err == nil {
			if _, err = os.ReadDir(mountPath); err == nil {
				m.logger.InfoContext(ctx, "mount is ready", "path", mountPath, "attempts", i+1)
				return nil
			}
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for mount ready: %w", ctx.Err())
		case <-time.After(sleepDuration):
		}
	}

	return fmt.Errorf("mount did not become ready within timeout")
}

func (m *Mounter) buildMountCmd(ctx context.Context, volumeID string, mountPath string) *exec.Cmd {
	args := []string{
		"--allow-other", "--allow-delete", "--allow-overwrite",
		"--file-mode", "0666", "--dir-mode", "0777",
		volumeID, mountPath,
	}

	envVars := m.buildEnvVars()

	cmd := exec.Command("mount-s3", args...)
	cmd.Env = envVars

	// On systemd hosts, wrap in systemd-run so the FUSE daemon survives runner restarts.
	if _, err := os.Stat("/run/systemd/system"); err == nil {
		sdArgs := []string{"--scope"}
		for _, env := range envVars {
			sdArgs = append(sdArgs, "--setenv="+env)
		}
		sdArgs = append(sdArgs, "--", "mount-s3")
		sdArgs = append(sdArgs, args...)
		cmd = exec.CommandContext(ctx, "systemd-run", sdArgs...)
	}

	cmd.Stderr = io.Writer(&log.ErrorLogWriter{})
	cmd.Stdout = io.Writer(&log.InfoLogWriter{})

	return cmd
}

func (m *Mounter) buildEnvVars() []string {
	var envVars []string
	if m.cfg.AWSEndpointUrl != "" {
		envVars = append(envVars, "AWS_ENDPOINT_URL="+m.cfg.AWSEndpointUrl)
	}
	if m.cfg.AWSAccessKeyId != "" {
		envVars = append(envVars, "AWS_ACCESS_KEY_ID="+m.cfg.AWSAccessKeyId)
	}
	if m.cfg.AWSSecretAccessKey != "" {
		envVars = append(envVars, "AWS_SECRET_ACCESS_KEY="+m.cfg.AWSSecretAccessKey)
	}
	if m.cfg.AWSRegion != "" {
		envVars = append(envVars, "AWS_REGION="+m.cfg.AWSRegion)
	}
	return envVars
}
