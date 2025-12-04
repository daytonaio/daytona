// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package mock

import (
	"context"
	"fmt"

	"github.com/daytonaio/mock-runner/pkg/api/dto"
	"github.com/daytonaio/mock-runner/pkg/models/enums"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	log "github.com/sirupsen/logrus"
)

// Create creates a mock sandbox (stores in memory, no real container)
func (m *MockClient) Create(ctx context.Context, sandboxDto dto.CreateSandboxDTO) (string, error) {
	log.Infof("Mock: Creating sandbox %s with snapshot %s", sandboxDto.Id, sandboxDto.Snapshot)

	// Check if sandbox already exists
	if existing, ok := m.getSandbox(sandboxDto.Id); ok {
		log.Infof("Mock: Sandbox %s already exists", sandboxDto.Id)
		// Return existing sandbox
		state, _ := m.DeduceSandboxState(ctx, sandboxDto.Id)
		if state == enums.SandboxStateStarted || state == enums.SandboxStatePullingSnapshot || state == enums.SandboxStateStarting {
			return existing.ID, nil
		}
		if state == enums.SandboxStateStopped || state == enums.SandboxStateCreating {
			err := m.Start(ctx, sandboxDto.Id, sandboxDto.Metadata)
			if err != nil {
				return "", err
			}
			return existing.ID, nil
		}
	}

	m.statesCache.SetSandboxState(ctx, sandboxDto.Id, enums.SandboxStateCreating)

	// Mock pulling image (just check if we track it)
	err := m.PullImage(ctx, sandboxDto.Snapshot, sandboxDto.Registry)
	if err != nil {
		return "", err
	}

	m.statesCache.SetSandboxState(ctx, sandboxDto.Id, enums.SandboxStateCreating)

	// Store sandbox info in memory
	sandbox := &SandboxInfo{
		ID:               sandboxDto.Id,
		UserId:           sandboxDto.UserId,
		Snapshot:         sandboxDto.Snapshot,
		OsUser:           sandboxDto.OsUser,
		CpuQuota:         sandboxDto.CpuQuota,
		GpuQuota:         sandboxDto.GpuQuota,
		MemoryQuota:      sandboxDto.MemoryQuota,
		StorageQuota:     sandboxDto.StorageQuota,
		Env:              sandboxDto.Env,
		Metadata:         sandboxDto.Metadata,
		NetworkBlockAll:  sandboxDto.NetworkBlockAll,
		NetworkAllowList: sandboxDto.NetworkAllowList,
	}
	m.setSandbox(sandbox)

	// Start the sandbox
	err = m.Start(ctx, sandboxDto.Id, sandboxDto.Metadata)
	if err != nil {
		return "", err
	}

	log.Infof("Mock: Sandbox %s created successfully", sandboxDto.Id)
	return sandboxDto.Id, nil
}

// Start starts a mock sandbox
func (m *MockClient) Start(ctx context.Context, containerId string, metadata map[string]string) error {
	log.Infof("Mock: Starting sandbox %s", containerId)

	m.statesCache.SetSandboxState(ctx, containerId, enums.SandboxStateStarting)

	// Ensure sandbox exists in memory
	sandbox, ok := m.getSandbox(containerId)
	if !ok {
		// Create a minimal sandbox entry if it doesn't exist
		sandbox = &SandboxInfo{
			ID:       containerId,
			Metadata: metadata,
		}
		m.setSandbox(sandbox)
	}

	// Ensure the toolbox container is running
	err := m.EnsureToolboxRunning(ctx)
	if err != nil {
		return fmt.Errorf("failed to ensure toolbox container is running: %w", err)
	}

	m.statesCache.SetSandboxState(ctx, containerId, enums.SandboxStateStarted)
	log.Infof("Mock: Sandbox %s started successfully", containerId)

	return nil
}

// Stop stops a mock sandbox
func (m *MockClient) Stop(ctx context.Context, containerId string) error {
	log.Infof("Mock: Stopping sandbox %s", containerId)

	state, _ := m.DeduceSandboxState(ctx, containerId)
	if state == enums.SandboxStateStopped {
		log.Debugf("Mock: Sandbox %s is already stopped", containerId)
		return nil
	}

	m.statesCache.SetSandboxState(ctx, containerId, enums.SandboxStateStopping)

	// Just update state - no real container to stop
	m.statesCache.SetSandboxState(ctx, containerId, enums.SandboxStateStopped)
	log.Infof("Mock: Sandbox %s stopped successfully", containerId)

	return nil
}

// Destroy destroys a mock sandbox
func (m *MockClient) Destroy(ctx context.Context, containerId string) error {
	log.Infof("Mock: Destroying sandbox %s", containerId)

	state, _ := m.DeduceSandboxState(ctx, containerId)
	if state == enums.SandboxStateDestroyed || state == enums.SandboxStateDestroying {
		log.Debugf("Mock: Sandbox %s is already destroyed or destroying", containerId)
		return nil
	}

	m.statesCache.SetSandboxState(ctx, containerId, enums.SandboxStateDestroying)

	// Remove sandbox from memory
	m.deleteSandbox(containerId)

	m.statesCache.SetSandboxState(ctx, containerId, enums.SandboxStateDestroyed)
	log.Infof("Mock: Sandbox %s destroyed successfully", containerId)

	return nil
}

// Resize resizes a mock sandbox (just update in-memory values)
func (m *MockClient) Resize(ctx context.Context, sandboxId string, sandboxDto dto.ResizeSandboxDTO) error {
	log.Infof("Mock: Resizing sandbox %s", sandboxId)

	m.statesCache.SetSandboxState(ctx, sandboxId, enums.SandboxStateResizing)

	sandbox, ok := m.getSandbox(sandboxId)
	if !ok {
		return fmt.Errorf("sandbox %s not found", sandboxId)
	}

	// Update resource values
	sandbox.CpuQuota = sandboxDto.Cpu
	sandbox.MemoryQuota = sandboxDto.Memory
	sandbox.GpuQuota = sandboxDto.Gpu
	m.setSandbox(sandbox)

	// Restore previous state (assume started)
	m.statesCache.SetSandboxState(ctx, sandboxId, enums.SandboxStateStarted)
	log.Infof("Mock: Sandbox %s resized successfully", sandboxId)

	return nil
}

// DeduceSandboxState returns the state from cache
func (m *MockClient) DeduceSandboxState(ctx context.Context, sandboxId string) (enums.SandboxState, error) {
	if sandboxId == "" {
		return enums.SandboxStateUnknown, nil
	}

	// Check if sandbox exists in memory
	_, ok := m.getSandbox(sandboxId)
	if !ok {
		return enums.SandboxStateDestroyed, nil
	}

	// Get state from cache
	cached, err := m.statesCache.Get(ctx, sandboxId)
	if err != nil || cached == nil {
		return enums.SandboxStateUnknown, nil
	}

	return cached.SandboxState, nil
}

// ContainerInspect returns mock container info with the toolbox container's IP
func (m *MockClient) ContainerInspect(ctx context.Context, containerId string) (container.InspectResponse, error) {
	sandbox, ok := m.getSandbox(containerId)
	if !ok {
		return container.InspectResponse{}, fmt.Errorf("sandbox %s not found", containerId)
	}

	// Get toolbox container IP
	toolboxIP := m.GetToolboxContainerIP()
	if toolboxIP == "" {
		toolboxIP = "127.0.0.1" // Fallback
	}

	state, _ := m.DeduceSandboxState(ctx, containerId)
	isRunning := state == enums.SandboxStateStarted

	// Build the response using proper struct initialization
	resp := container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{
			ID: containerId,
			State: &container.State{
				Running: isRunning,
				Status:  stateToDockerStatus(state),
			},
		},
		Config: &container.Config{
			WorkingDir: "/home/daytona",
			Env:        mapToEnvSlice(sandbox.Env),
		},
		NetworkSettings: &container.NetworkSettings{
			Networks: map[string]*network.EndpointSettings{
				"bridge": {
					IPAddress: toolboxIP,
				},
			},
		},
	}

	return resp, nil
}

// StartBackupCreate creates a mock backup (no-op but tracks state)
func (m *MockClient) StartBackupCreate(ctx context.Context, containerId string, backupDto dto.CreateBackupDTO) error {
	log.Infof("Mock: Creating backup for sandbox %s", containerId)

	m.statesCache.SetBackupState(ctx, containerId, enums.BackupStateInProgress, nil)

	// Mock: Just mark as completed after a brief moment
	go func() {
		m.statesCache.SetBackupState(ctx, containerId, enums.BackupStateCompleted, nil)
		log.Infof("Mock: Backup for sandbox %s completed", containerId)
	}()

	return nil
}

// Helper functions

func stateToDockerStatus(state enums.SandboxState) string {
	switch state {
	case enums.SandboxStateStarted:
		return "running"
	case enums.SandboxStateStopped:
		return "exited"
	case enums.SandboxStateCreating:
		return "created"
	case enums.SandboxStateDestroyed:
		return "dead"
	default:
		return "unknown"
	}
}

func mapToEnvSlice(env map[string]string) []string {
	if env == nil {
		return nil
	}
	result := make([]string, 0, len(env))
	for k, v := range env {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}
