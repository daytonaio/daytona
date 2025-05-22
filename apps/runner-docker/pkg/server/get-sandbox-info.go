// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package server

import (
	"context"
	"fmt"
	"strings"

	pb "github.com/daytonaio/runner/proto"

	"github.com/daytonaio/runner-docker/pkg/models/enums"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func (s *RunnerServer) GetSandboxInfo(ctx context.Context, req *pb.GetSandboxInfoRequest) (*pb.GetSandboxInfoResponse, error) {
	sandboxState, err := s.getSandboxState(ctx, req.SandboxId)
	if err == nil {
		s.cache.SetSandboxState(ctx, req.SandboxId, sandboxState)
	}

	data := s.cache.Get(ctx, req.SandboxId)

	if data == nil {
		return &pb.GetSandboxInfoResponse{
			State:         enums.SandboxStateUnknown.String(),
			SnapshotState: enums.SnapshotStateNone.String(),
		}, nil
	}

	return &pb.GetSandboxInfoResponse{
		State:         data.SandboxState.String(),
		SnapshotState: data.SnapshotState.String(),
	}, nil
}

func (s *RunnerServer) getSandboxState(ctx context.Context, sandboxId string) (enums.SandboxState, error) {
	if sandboxId == "" {
		return enums.SandboxStateUnknown, nil
	}

	container, err := s.dockerClient.ContainerInspect(ctx, sandboxId)
	if err != nil {
		if client.IsErrNotFound(err) {
			return enums.SandboxStateDestroyed, nil
		}
		return enums.SandboxStateError, fmt.Errorf("failed to inspect container: %w", err)
	}

	switch container.State.Status {
	case "created":
		return enums.SandboxStateCreating, nil

	case "running":
		if s.isContainerPullingImage(container.ID) {
			return enums.SandboxStatePullingImage, nil
		}
		return enums.SandboxStateStarted, nil

	case "paused":
		return enums.SandboxStateStopped, nil

	case "restarting":
		return enums.SandboxStateStarting, nil

	case "removing":
		return enums.SandboxStateDestroying, nil

	case "exited":
		if container.State.ExitCode == 0 || container.State.ExitCode == 137 || container.State.ExitCode == 143 {
			return enums.SandboxStateStopped, nil
		}

		return enums.SandboxStateError, fmt.Errorf("container exited with code %d, reason: %s", container.State.ExitCode, container.State.Error)

	case "dead":
		return enums.SandboxStateDestroyed, nil

	default:
		return enums.SandboxStateUnknown, nil
	}
}

// isContainerPullingImage checks if the container is still in image pulling phase
func (s *RunnerServer) isContainerPullingImage(containerName string) bool {
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       "10", // Look at last 10 lines
	}

	logs, err := s.dockerClient.ContainerLogs(context.Background(), containerName, options)
	if err != nil {
		return false
	}
	defer logs.Close()

	// Read logs and check for pull messages
	buf := make([]byte, 1024)
	n, _ := logs.Read(buf)
	logContent := string(buf[:n])

	return strings.Contains(logContent, "Pulling from") ||
		strings.Contains(logContent, "Downloading") ||
		strings.Contains(logContent, "Extracting")
}
