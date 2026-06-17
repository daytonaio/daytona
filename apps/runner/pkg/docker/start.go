// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/daytonaio/common-go/pkg/timer"
	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/pkg/stdcopy"
)

func (d *DockerClient) Start(ctx context.Context, containerId string, authToken *string, metadata map[string]string) (*container.InspectResponse, string, error) {
	defer timer.Timer()()

	// Cancel a backup if it's already in progress
	backup_context, ok := backup_context_map.Get(containerId)
	if ok {
		backup_context.cancel()
	}

	c, err := d.ContainerInspect(ctx, containerId)
	if err != nil {
		return nil, "", err
	}

	if c.State.Running {
		containerIP := GetContainerIpAddress(ctx, c)
		if containerIP == "" {
			return nil, "", errors.New("sandbox IP not found? Is the sandbox started?")
		}

		if isAndroidDeviceContainer(c) {
			if err := d.waitForAdbRunning(ctx, containerIP); err != nil {
				return nil, "", err
			}
			return c, "", nil
		}

		daemonVersion, err := d.waitForDaemonRunning(ctx, containerIP, authToken)
		if err != nil {
			return nil, "", err
		}

		return c, daemonVersion, nil
	}

	// Re-establish FUSE mounts that may have died since the container was last running.
	if volumesJSON, ok := metadata["volumes"]; ok {
		var volumes []dto.VolumeDTO
		if err := json.Unmarshal([]byte(volumesJSON), &volumes); err == nil && len(volumes) > 0 {
			mounter := d.resolveVolumeMounter(metadata)
			_, err = d.getVolumesMountPathBinds(ctx, volumes, mounter)
			if err != nil {
				d.logger.ErrorContext(ctx, "Failed to ensure volume FUSE mounts", "error", err)
			}
		}
	}

	err = d.apiClient.ContainerStart(ctx, containerId, container.StartOptions{})
	if err != nil {
		return nil, "", err
	}

	// make sure container is running
	runningContainer, err := d.waitForContainerRunning(ctx, containerId)
	if err != nil {
		return nil, "", err
	}

	containerIP := GetContainerIpAddress(ctx, runningContainer)
	if containerIP == "" {
		return nil, "", errors.New("sandbox IP not found? Is the sandbox started?")
	}

	// Android-device sandboxes do not run the daytona daemon. Readiness is signaled by
	// the ADB port accepting TCP connections inside the container.
	if isAndroidDeviceContainer(runningContainer) {
		if err := d.waitForAdbRunning(ctx, containerIP); err != nil {
			return nil, "", err
		}

		if metadata["limitNetworkEgress"] == "true" {
			go func() {
				containerShortId := c.ID[:12]
				if err := d.netRulesManager.SetNetworkLimiter(containerShortId, containerIP); err != nil {
					d.logger.ErrorContext(ctx, "Failed to set network limiter", "error", err)
				}
			}()
		}

		return runningContainer, "", nil
	}

	if !slices.Equal(c.Config.Entrypoint, strslice.StrSlice{common.DAEMON_PATH}) {
		processesCtx := context.Background()
		go func() {
			if err := d.startDaytonaDaemon(processesCtx, containerId, c.Config.WorkingDir); err != nil {
				d.logger.ErrorContext(ctx, "Failed to start Daytona daemon", "error", err)
			}
		}()
	}

	// If daemon is the sandbox entrypoint (common.DAEMON_PATH), it is started as part of the sandbox;
	// Otherwise, the daemon is started separately above.
	// In either case, we wait for it here.
	daemonVersion, err := d.waitForDaemonRunning(ctx, containerIP, authToken)
	if err != nil {
		return nil, "", err
	}

	if metadata["limitNetworkEgress"] == "true" {
		go func() {
			containerShortId := c.ID[:12]
			err = d.netRulesManager.SetNetworkLimiter(containerShortId, containerIP)
			if err != nil {
				d.logger.ErrorContext(ctx, "Failed to set network limiter", "error", err)
			}
		}()
	}

	return runningContainer, daemonVersion, nil
}

func (d *DockerClient) waitForContainerRunning(ctx context.Context, containerId string) (*container.InspectResponse, error) {
	defer timer.Timer()()

	timeout := time.Duration(d.sandboxStartTimeoutSec) * time.Second
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return nil, errors.New("timeout waiting for the sandbox to start - please ensure that your entrypoint is long-running")
		case <-ticker.C:
			c, err := d.ContainerInspect(timeoutCtx, containerId)
			if err != nil {
				return nil, err
			}

			if c.State.Running {
				return c, nil
			}

			// Detect a container that exited before becoming healthy.
			// Otherwise we'd poll until the start timeout and report a
			// generic "entrypoint must be long-running" message that hides
			// the real cause (e.g. a daemon-side volume mount failure).
			if c.State.Status == container.StateExited && c.State.ExitCode != 0 {
				return nil, d.formatEarlyExitError(ctx, containerId, c)
			}
		}
	}
}

// formatEarlyExitError builds a user-facing error for a container that exited
// before reaching Running, combining the docker exit code/reason with the
// tail of its logs so the user gets actionable signal beyond an exit code.
func (d *DockerClient) formatEarlyExitError(ctx context.Context, containerId string, c *container.InspectResponse) error {
	tail := d.tailContainerLogs(ctx, containerId, 25)
	tail = strings.TrimSpace(tail)

	// Keep the message short when there are no logs, verbose when there are.
	parts := []string{
		fmt.Sprintf("sandbox exited prematurely with code %d", c.State.ExitCode),
	}
	if c.State.Error != "" {
		parts = append(parts, fmt.Sprintf("docker reason: %s", c.State.Error))
	}
	if tail != "" {
		parts = append(parts, fmt.Sprintf("last container logs:\n%s", tail))
	} else {
		parts = append(parts, "no container logs available — check the runner host's docker logs")
	}
	return errors.New(strings.Join(parts, "; "))
}

// tailContainerLogs returns up to `tail` lines of combined stdout/stderr,
// best-effort. Errors are swallowed so a missing tail can't mask the original
// failure.
func (d *DockerClient) tailContainerLogs(ctx context.Context, containerId string, tail int) string {
	rdr, err := d.apiClient.ContainerLogs(ctx, containerId, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       fmt.Sprintf("%d", tail),
	})
	if err != nil {
		return ""
	}
	defer rdr.Close()

	var stdout, stderr bytes.Buffer
	if _, err := stdcopy.StdCopy(&stdout, &stderr, rdr); err != nil {
		// Keep whatever bytes were written before the error.
	}

	var combined strings.Builder
	if stderr.Len() > 0 {
		combined.WriteString(strings.TrimRight(stderr.String(), "\n"))
	}
	if stdout.Len() > 0 {
		if combined.Len() > 0 {
			combined.WriteString("\n")
		}
		combined.WriteString(strings.TrimRight(stdout.String(), "\n"))
	}
	return combined.String()
}
