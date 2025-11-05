// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"time"

	common_daemon "github.com/daytonaio/common-go/pkg/daemon"
	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/common-go/pkg/timer"
	"github.com/docker/docker/api/types/container"

	log "github.com/sirupsen/logrus"
)

func (d *DockerClient) startDaytonaDaemon(ctx context.Context, containerId string, workDir string) error {
	defer timer.Timer()()

	daemonCmd := "/usr/local/bin/daytona"
	if workDir == "" {
		workDir = common_daemon.UseUserHomeAsWorkDir
	}
	daemonCmd = fmt.Sprintf("%s --work-dir %s", daemonCmd, workDir)

	execOptions := container.ExecOptions{
		Cmd:          []string{"sh", "-c", daemonCmd},
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
	}

	execStartOptions := container.ExecStartOptions{
		Detach: false,
	}

	result, err := d.execSync(ctx, containerId, execOptions, execStartOptions)
	if err != nil {
		log.Errorf("Error starting Daytona daemon: %s", err.Error())
		return nil
	}

	if result.ExitCode != 0 && result.StdErr != "" {
		log.Errorf("Error starting Daytona daemon: %s", string(result.StdErr))
		return nil
	}

	return nil
}

// getDaemonWrapperEntrypoint creates an entrypoint command that:
// 1. Executes the snapshot entrypoint if provided (in background)
// 2. Execs into the Daytona daemon so it becomes PID1
// This ensures the daemon is PID1 while still executing the snapshot entrypoint
func (d *DockerClient) getDaemonWrapperEntrypoint(workDir string) []string {
	// Build the wrapper script as a single sh -c command
	// This script:
	// - Executes the snapshot entrypoint in background if DAYTONA_SNAPSHOT_ENTRYPOINT is set
	// - Execs into the daemon, making it PID1
	wrapperScript := fmt.Sprintf(`#!/bin/sh
set -e

# Execute snapshot entrypoint if provided (run in background)
# Note: We don't use set -e for this part to ensure daemon always starts
if [ -n "$DAYTONA_SNAPSHOT_ENTRYPOINT" ]; then
	# Parse JSON array and convert to shell command
	# Example: ["sleep", "infinity"] -> sleep infinity
	# Remove brackets, quotes, and commas, then reconstruct command
	ENTRYPOINT_CMD=$(echo "$DAYTONA_SNAPSHOT_ENTRYPOINT" | sed 's/^\[//;s/\]$//;s/"//g; s/,/ /g' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
	if [ -n "$ENTRYPOINT_CMD" ]; then
		# Execute the entrypoint in background
		# This allows it to run while daemon becomes PID1
		# If entrypoint fails, we still continue to start daemon
		(sh -c "$ENTRYPOINT_CMD" || true) &
	fi
fi

# Exec into daemon so it becomes PID1
# This replaces the shell process (PID1) with the daemon process
exec /usr/local/bin/daytona --work-dir %s
`, workDir)

	return []string{"sh", "-c", wrapperScript}
}

func (d *DockerClient) waitForDaemonRunning(ctx context.Context, containerIP string, timeout time.Duration) error {
	defer timer.Timer()()

	// Build the target URL
	targetURL := fmt.Sprintf("http://%s:2280/version", containerIP)
	target, err := url.Parse(targetURL)
	if err != nil {
		return common_errors.NewBadRequestError(fmt.Errorf("failed to parse target URL: %w", err))
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("timeout waiting for daemon to start")
		default:
			conn, err := net.DialTimeout("tcp", target.Host, 1*time.Second)
			if err != nil {
				time.Sleep(5 * time.Millisecond)
				continue
			}
			conn.Close()
			return nil
		}
	}
}
