// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	apiclient "github.com/daytonaio/apiclient"
	runnerapiclient "github.com/daytonaio/runner/pkg/apiclient"
	"github.com/daytonaio/runner/pkg/docker"
	"github.com/daytonaio/runner/pkg/models/enums"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

type SandboxSyncServiceConfig struct {
	Docker   *docker.DockerClient
	Interval time.Duration
}

type SandboxSyncService struct {
	docker   *docker.DockerClient
	interval time.Duration
	client   *apiclient.APIClient
}

func NewSandboxSyncService(config SandboxSyncServiceConfig) *SandboxSyncService {
	return &SandboxSyncService{
		docker:   config.Docker,
		interval: config.Interval,
	}
}

func (s *SandboxSyncService) GetLocalContainerStates(ctx context.Context) (map[string]enums.SandboxState, error) {
	containers, err := s.docker.ApiClient().ContainerList(ctx, container.ListOptions{
		All: true, // Include stopped containers
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	containerStates := make(map[string]enums.SandboxState)

	for _, container := range containers {
		// Extract sandbox ID from container name or labels
		sandboxId := s.extractSandboxId(container)
		if sandboxId == "" {
			continue // Skip non-sandbox containers
		}

		// Get the current state of this container
		state, err := s.docker.DeduceSandboxState(ctx, sandboxId)
		if err != nil {
			slog.DebugContext(ctx, "Failed to deduce state for sandbox", "sandboxId", sandboxId, "error", err)
			continue
		}

		containerStates[sandboxId] = state
	}

	return containerStates, nil
}

func (s *SandboxSyncService) GetRemoteSandboxStates(ctx context.Context) (map[string]apiclient.SandboxState, error) {
	if s.client == nil {
		client, err := runnerapiclient.GetApiClient()
		if err != nil {
			return nil, fmt.Errorf("failed to get API client: %w", err)
		}
		s.client = client
	}
	sandboxes, _, err := s.client.SandboxAPI.GetSandboxesForRunner(ctx).
		States(string(apiclient.SANDBOXSTATE_STARTED)).SkipReconcilingSandboxes(true).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get sandboxes from API: %w", err)
	}

	remoteSandboxes := make(map[string]apiclient.SandboxState)
	for _, sandbox := range sandboxes {
		if sandbox.Id != "" {
			remoteSandboxes[sandbox.Id] = *sandbox.State
		}
	}

	return remoteSandboxes, nil
}

func (s *SandboxSyncService) SyncSandboxState(ctx context.Context, sandboxId string, localState enums.SandboxState) error {
	_, err := s.client.SandboxAPI.UpdateSandboxState(ctx, sandboxId).UpdateSandboxStateDto(*apiclient.NewUpdateSandboxStateDto(
		string(s.convertToApiState(localState)),
	)).Execute()
	if err != nil {
		return fmt.Errorf("failed to get sandbox %s: %w", sandboxId, err)
	}

	return nil
}

func (s *SandboxSyncService) PerformSync(ctx context.Context) error {
	localStates, err := s.GetLocalContainerStates(ctx)
	if err != nil {
		return fmt.Errorf("failed to get local container states: %w", err)
	}

	remoteStates, err := s.GetRemoteSandboxStates(ctx)
	if err != nil {
		return fmt.Errorf("failed to get remote sandbox states: %w", err)
	}

	// Compare states and sync differences
	syncCount := 0
	for sandboxId, localState := range localStates {
		remoteState, exists := remoteStates[sandboxId]
		if !exists {
			continue
		}

		// Convert remote state to local state format for comparison
		convertedRemoteState := s.convertFromApiState(remoteState)

		if localState != convertedRemoteState {
			slog.InfoContext(ctx, "State mismatch for sandbox", "sandboxId", sandboxId, "localState", localState, "remoteState", convertedRemoteState)

			err := s.SyncSandboxState(ctx, sandboxId, localState)
			if err != nil {
				slog.ErrorContext(ctx, "Failed to sync state for sandbox", "sandboxId", sandboxId, "error", err)
				continue
			}
			syncCount++
		}
	}

	if syncCount > 0 {
		slog.InfoContext(ctx, "Synchronized sandbox states", "syncCount", syncCount)
	}

	return nil
}

// StartSyncProcess starts a background goroutine that synchronizes sandbox states
func (s *SandboxSyncService) StartSyncProcess(ctx context.Context) {
	slog.InfoContext(ctx, "Starting sandbox sync process")
	go func() {
		// Perform initial sync
		err := s.PerformSync(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to perform initial sync", "error", err)
		}

		// Set up ticker for periodic sync
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				err := s.PerformSync(ctx)
				if err != nil {
					slog.ErrorContext(ctx, "Failed to perform sync", "error", err)
				}
			case <-ctx.Done():
				slog.InfoContext(ctx, "Sandbox sync service stopped")
				return
			}
		}
	}()
}

func (s *SandboxSyncService) extractSandboxId(container types.Container) string {
	if len(container.Names) > 0 && len(container.Names[0]) > 1 {
		name := container.Names[0][1:] // Remove leading "/"
		return name
	}

	return ""
}

func (s *SandboxSyncService) convertToApiState(localState enums.SandboxState) apiclient.SandboxState {
	switch localState {
	case enums.SandboxStateCreating:
		return apiclient.SANDBOXSTATE_CREATING
	case enums.SandboxStateRestoring:
		return apiclient.SANDBOXSTATE_RESTORING
	case enums.SandboxStateDestroyed:
		return apiclient.SANDBOXSTATE_DESTROYED
	case enums.SandboxStateDestroying:
		return apiclient.SANDBOXSTATE_DESTROYING
	case enums.SandboxStateStarted:
		return apiclient.SANDBOXSTATE_STARTED
	case enums.SandboxStateStopped:
		return apiclient.SANDBOXSTATE_STOPPED
	case enums.SandboxStateStarting:
		return apiclient.SANDBOXSTATE_STARTING
	case enums.SandboxStateStopping:
		return apiclient.SANDBOXSTATE_STOPPING
	case enums.SandboxStateError:
		return apiclient.SANDBOXSTATE_ERROR
	case enums.SandboxStatePullingSnapshot:
		return apiclient.SANDBOXSTATE_PULLING_SNAPSHOT
	default:
		return apiclient.SANDBOXSTATE_UNKNOWN
	}
}

func (s *SandboxSyncService) convertFromApiState(apiState apiclient.SandboxState) enums.SandboxState {
	switch apiState {
	case apiclient.SANDBOXSTATE_CREATING:
		return enums.SandboxStateCreating
	case apiclient.SANDBOXSTATE_RESTORING:
		return enums.SandboxStateRestoring
	case apiclient.SANDBOXSTATE_DESTROYED:
		return enums.SandboxStateDestroyed
	case apiclient.SANDBOXSTATE_DESTROYING:
		return enums.SandboxStateDestroying
	case apiclient.SANDBOXSTATE_STARTED:
		return enums.SandboxStateStarted
	case apiclient.SANDBOXSTATE_STOPPED:
		return enums.SandboxStateStopped
	case apiclient.SANDBOXSTATE_STARTING:
		return enums.SandboxStateStarting
	case apiclient.SANDBOXSTATE_STOPPING:
		return enums.SandboxStateStopping
	case apiclient.SANDBOXSTATE_ERROR:
		return enums.SandboxStateError
	case apiclient.SANDBOXSTATE_PULLING_SNAPSHOT:
		return enums.SandboxStatePullingSnapshot
	default:
		return enums.SandboxStateUnknown
	}
}
