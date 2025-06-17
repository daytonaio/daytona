// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"fmt"

	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
	"github.com/daytonaio/runner-docker/pkg/common"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/errdefs"
)

func (s *SandboxService) RemoveDestroyedSandbox(ctx context.Context, req *pb.RemoveDestroyedSandboxRequest) (*pb.RemoveDestroyedSandboxResponse, error) {
	// Check if container exists and is in destroyed state
	state, err := s.getSandboxState(ctx, req.GetSandboxId())
	if err != nil {
		return nil, err
	}

	if state != pb.SandboxState_SANDBOX_STATE_DESTROYED {
		return nil, fmt.Errorf("sandbox %s is not in destroyed state", req.GetSandboxId())
	}

	// Remove the container
	err = s.dockerClient.ContainerRemove(ctx, req.GetSandboxId(), container.RemoveOptions{
		Force: true,
	})
	if err != nil {
		if errdefs.IsNotFound(err) {
			s.cache.SetSandboxState(ctx, req.GetSandboxId(), pb.SandboxState_SANDBOX_STATE_DESTROYED)

			return &pb.RemoveDestroyedSandboxResponse{
				Message: fmt.Sprintf("Destroyed sandbox %s already removed", req.GetSandboxId()),
			}, nil
		}

		return nil, common.MapDockerError(err)
	}

	s.log.Info("Destroyed sandbox removed successfully", "sandboxId", req.GetSandboxId())

	return &pb.RemoveDestroyedSandboxResponse{
		Message: fmt.Sprintf("Destroyed sandbox %s removed", req.GetSandboxId()),
	}, nil
}
