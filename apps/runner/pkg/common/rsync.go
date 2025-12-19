// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

// RsyncCopy copies files from srcPath to destPath using rsync with full attribute preservation.
// It uses rsync with -aAX flags to preserve permissions, ownership, timestamps, symlinks,
// devices, ACLs, and extended attributes.
//
// The timeout parameter specifies how long to wait for the rsync operation to complete.
// Trailing slashes are automatically added to paths to ensure contents are copied, not directories.
func RsyncCopy(ctx context.Context, srcPath, destPath string) error {
	log.Debugf("rsync copy from %s to %s", srcPath, destPath)

	// Use rsync with -aAX flags:
	// -a = archive mode (preserves permissions, ownership, timestamps, symlinks, devices)
	// -A = preserve ACLs
	// -X = preserve extended attributes (xattrs)
	// Trailing slashes ensure we copy contents, not the directory itself
	src := filepath.Clean(srcPath) + "/"
	dest := filepath.Clean(destPath) + "/"
	rsyncCmd := exec.CommandContext(ctx, "rsync", "-aAX", src, dest)

	var rsyncOut strings.Builder
	var rsyncErr strings.Builder
	rsyncCmd.Stdout = &rsyncOut
	rsyncCmd.Stderr = &rsyncErr

	log.Debug("Starting rsync...")
	if err := rsyncCmd.Run(); err != nil {
		if errMsg := rsyncErr.String(); errMsg != "" {
			log.Errorf("rsync stderr: %s", errMsg)
		}
		return fmt.Errorf("rsync failed: %w", err)
	}

	if outMsg := rsyncOut.String(); outMsg != "" {
		log.Debugf("rsync output: %s", outMsg)
	}

	log.Info("Successfully completed rsync copy")
	return nil
}
