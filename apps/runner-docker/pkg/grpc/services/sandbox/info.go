// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"fmt"
	"strings"

	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *SandboxService) SandboxInfo(ctx context.Context, req *pb.SandboxInfoRequest) (*pb.SandboxInfoResponse, error) {
	sandboxState, err := s.getSandboxState(ctx, req.GetSandboxId())
	if err == nil {
		s.cache.SetSandboxState(ctx, req.GetSandboxId(), sandboxState)
	}

	data := s.cache.Get(ctx, req.GetSandboxId())

	if data == nil {
		return &pb.SandboxInfoResponse{
			State:       pb.SandboxState_SANDBOX_STATE_UNSPECIFIED,
			BackupState: pb.BackupState_BACKUP_STATE_UNSPECIFIED,
		}, nil
	}

	return &pb.SandboxInfoResponse{
		State:       data.SandboxState,
		BackupState: data.BackupState,
	}, nil
}

func (s *SandboxService) getSandboxState(ctx context.Context, sandboxId string) (pb.SandboxState, error) {
	if sandboxId == "" {
		return pb.SandboxState_SANDBOX_STATE_UNSPECIFIED, nil
	}

	container, err := s.dockerClient.ContainerInspect(ctx, sandboxId)
	if err != nil {
		if client.IsErrNotFound(err) {
			return pb.SandboxState_SANDBOX_STATE_DESTROYED, nil
		}
		return pb.SandboxState_SANDBOX_STATE_ERROR, status.Error(codes.Internal, fmt.Errorf("failed to inspect container: %w", err).Error())
	}

	switch container.State.Status {
	case "created":
		return pb.SandboxState_SANDBOX_STATE_CREATING, nil

	case "running":
		if s.isContainerPullingImage(container.ID) {
			return pb.SandboxState_SANDBOX_STATE_PULLING_SNAPSHOT, nil
		}
		return pb.SandboxState_SANDBOX_STATE_STARTED, nil

	case "paused":
		return pb.SandboxState_SANDBOX_STATE_STOPPED, nil

	case "restarting":
		return pb.SandboxState_SANDBOX_STATE_STARTING, nil

	case "removing":
		return pb.SandboxState_SANDBOX_STATE_DESTROYING, nil

	case "exited":
		if container.State.ExitCode == 0 || container.State.ExitCode == 137 || container.State.ExitCode == 143 {
			return pb.SandboxState_SANDBOX_STATE_STOPPED, nil
		}

		return pb.SandboxState_SANDBOX_STATE_ERROR, status.Error(codes.Internal, fmt.Errorf("container exited with code %d, reason: %s", container.State.ExitCode, container.State.Error).Error())

	case "dead":
		return pb.SandboxState_SANDBOX_STATE_DESTROYED, nil

	default:
		return pb.SandboxState_SANDBOX_STATE_UNSPECIFIED, nil
	}
}

// isContainerPullingImage checks if the container is still in image pulling phase
func (s *SandboxService) isContainerPullingImage(containerName string) bool {
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
