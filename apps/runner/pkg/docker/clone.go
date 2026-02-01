// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/daytonaio/common-go/pkg/timer"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/daytonaio/runner/pkg/models/enums"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"

	log "github.com/sirupsen/logrus"
)

// Type aliases for Docker API types
type CommitOptions = container.CommitOptions
type ImageRemoveOptions = image.RemoveOptions
type StartOptions = container.StartOptions

// CloneSandboxInfo holds information about a cloned sandbox
type CloneSandboxInfo struct {
	Id    string
	State enums.SandboxState
}

// CloneSandbox creates an independent copy of a sandbox with flattened filesystem
// Unlike Docker's typical layered approach, this commits the container to a new image
// and creates a fresh container from it, ensuring no dependency on the source container
func (d *DockerClient) CloneSandbox(ctx context.Context, sourceSandboxId string, newSandboxId string) (*CloneSandboxInfo, error) {
	defer timer.Timer()()

	startTime := time.Now()
	defer func() {
		obs, err := common.ContainerOperationDuration.GetMetricWithLabelValues("clone")
		if err == nil {
			obs.Observe(time.Since(startTime).Seconds())
		}
	}()

	log.Infof("Cloning sandbox %s to %s", sourceSandboxId, newSandboxId)

	// Step 1: Check if source container exists
	sourceContainer, err := d.ContainerInspect(ctx, sourceSandboxId)
	if err != nil {
		return nil, fmt.Errorf("source sandbox not found: %w", err)
	}

	// Step 2: If source container is running, pause it for consistent snapshot
	wasRunning := sourceContainer.State.Running
	if wasRunning {
		log.Infof("Pausing source container %s for clone", sourceSandboxId)
		if err := d.apiClient.ContainerPause(ctx, sourceSandboxId); err != nil {
			return nil, fmt.Errorf("failed to pause source container: %w", err)
		}
		// Ensure we unpause on any error or success
		defer func() {
			log.Infof("Unpausing source container %s after clone", sourceSandboxId)
			if err := d.apiClient.ContainerUnpause(ctx, sourceSandboxId); err != nil {
				log.Warnf("Failed to unpause source container: %v", err)
			}
		}()
	}

	// Step 3: Commit the container to a temporary image
	// This captures the complete filesystem state
	tempImageName := fmt.Sprintf("daytona-clone-temp:%s", newSandboxId)
	log.Infof("Committing container %s to temporary image %s", sourceSandboxId, tempImageName)

	commitResp, err := d.apiClient.ContainerCommit(ctx, sourceSandboxId, CommitOptions{
		Reference: tempImageName,
		Comment:   fmt.Sprintf("Clone of sandbox %s", sourceSandboxId),
		Author:    "daytona",
		Pause:     false, // Already paused if needed
	})
	if err != nil {
		return nil, fmt.Errorf("failed to commit container: %w", err)
	}

	log.Infof("Created temporary image %s (ID: %s)", tempImageName, commitResp.ID)

	// Ensure cleanup of temporary image
	defer func() {
		log.Infof("Cleaning up temporary image %s", tempImageName)
		if _, err := d.apiClient.ImageRemove(context.Background(), tempImageName, ImageRemoveOptions{Force: true}); err != nil {
			log.Warnf("Failed to remove temporary image %s: %v", tempImageName, err)
		}
	}()

	// Step 4: Create new container from the committed image
	log.Infof("Creating new container %s from temporary image", newSandboxId)

	// Get the original container's config and modify it for the clone
	containerConfig := sourceContainer.Config
	hostConfig := sourceContainer.HostConfig

	// Update hostname and environment variables for the new sandbox
	containerConfig.Hostname = newSandboxId
	containerConfig.Image = tempImageName

	// Update DAYTONA_SANDBOX_ID in environment
	for i, env := range containerConfig.Env {
		if len(env) > 17 && env[:18] == "DAYTONA_SANDBOX_ID" {
			containerConfig.Env[i] = "DAYTONA_SANDBOX_ID=" + newSandboxId
			break
		}
	}

	// Create the new container
	_, err = d.apiClient.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, newSandboxId)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloned container: %w", err)
	}

	// Step 5: Start the new container
	log.Infof("Starting cloned container %s", newSandboxId)
	if err := d.apiClient.ContainerStart(ctx, newSandboxId, StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start cloned container: %w", err)
	}

	// Step 6: Wait for daemon to be ready
	if err := d.waitForDaemonReady(ctx, newSandboxId); err != nil {
		log.Warnf("Daemon health check failed for %s: %v (continuing anyway)", newSandboxId, err)
	}

	log.Infof("Clone completed: %s -> %s", sourceSandboxId, newSandboxId)

	return &CloneSandboxInfo{
		Id:    newSandboxId,
		State: enums.SandboxStateStarted,
	}, nil
}

// waitForDaemonReady waits for the daemon to be reachable in the cloned container
func (d *DockerClient) waitForDaemonReady(ctx context.Context, sandboxId string) error {
	timeout := time.Duration(d.daemonStartTimeoutSec) * time.Second
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		_, err := d.GetDaemonVersion(ctx, sandboxId)
		if err == nil {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("daemon not ready after %v", timeout)
}
