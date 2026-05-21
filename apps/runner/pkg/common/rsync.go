// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"
)

// RsyncCopy copies files from srcPath to destPath using rsync with full attribute preservation.
// It uses rsync with -aAXS flags to preserve permissions, ownership, timestamps, symlinks,
// devices, ACLs, extended attributes, and sparse file efficiency.
//
// Timeouts are controlled via the passed context.Context deadline (e.g. context.WithTimeout).
// Trailing slashes are automatically added to paths to ensure contents are copied, not directories.
func RsyncCopy(ctx context.Context, logger *slog.Logger, srcPath, destPath string) error {
	logger.DebugContext(ctx, "rsync copy", "source", srcPath, "destination", destPath)

	src := filepath.Clean(srcPath) + "/"
	dest := filepath.Clean(destPath) + "/"
	rsyncCmd := exec.CommandContext(ctx, "rsync", "-aAXS", src, dest)

	var rsyncOut strings.Builder
	var rsyncErr strings.Builder
	rsyncCmd.Stdout = &rsyncOut
	rsyncCmd.Stderr = &rsyncErr

	logger.DebugContext(ctx, "Starting rsync...")
	if err := rsyncCmd.Run(); err != nil {
		stderr := strings.TrimSpace(rsyncErr.String())
		logger.ErrorContext(ctx, "rsync failed", "error", err, "stderr", stderr)
		return rsyncError(ctx, err, stderr)
	}

	if outMsg := rsyncOut.String(); outMsg != "" {
		logger.DebugContext(ctx, "rsync output", "output", outMsg)
	}

	logger.InfoContext(ctx, "Successfully completed rsync copy")
	return nil
}

// rsyncError maps an rsync failure to a clean, user-facing message so the raw
// "exit status N" and stderr (already logged) never reach end users. The out-of-space
// wording is kept compatible with the recoverable patterns in recovery.go.
func rsyncError(ctx context.Context, runErr error, stderr string) error {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return errors.New("data copy timed out before completing")
	}

	stderrLower := strings.ToLower(stderr)
	if strings.Contains(stderrLower, "no space left on device") ||
		strings.Contains(stderrLower, "disk quota exceeded") {
		return errors.New("data copy failed: no space left on device")
	}

	var exitErr *exec.ExitError
	if !errors.As(runErr, &exitErr) {
		return errors.New("data copy failed")
	}

	switch exitErr.ExitCode() {
	case 11:
		return errors.New("data copy failed due to a file I/O error")
	case 12:
		return errors.New("data copy failed due to an rsync protocol error")
	case 23, 24:
		return errors.New("data copy did not complete: some files could not be transferred")
	case 30:
		return errors.New("data copy timed out before completing")
	default:
		return fmt.Errorf("data copy failed unexpectedly (rsync exit code %d)", exitErr.ExitCode())
	}
}
