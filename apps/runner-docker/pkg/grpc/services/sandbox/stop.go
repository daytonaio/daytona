// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"fmt"
	"time"

	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
	"github.com/daytonaio/runner-docker/pkg/common"
	"github.com/docker/docker/api/types/container"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *SandboxService) StopSandbox(ctx context.Context, req *pb.StopSandboxRequest) (*pb.StopSandboxResponse, error) {
	s.cache.SetSandboxState(ctx, req.GetSandboxId(), pb.SandboxState_SANDBOX_STATE_STOPPING)

	err := s.dockerClient.ContainerStop(ctx, req.GetSandboxId(), container.StopOptions{
		Signal: "SIGKILL",
	})
	if err != nil {
		return nil, common.MapDockerError(err)
	}

	err = s.waitForContainerStopped(ctx, req.GetSandboxId(), 10*time.Second)
	if err != nil {
		return nil, err
	}

	s.cache.SetSandboxState(ctx, req.GetSandboxId(), pb.SandboxState_SANDBOX_STATE_STOPPED)

	return &pb.StopSandboxResponse{
		Message: fmt.Sprintf("Sandbox %s stopped", req.GetSandboxId()),
	}, nil
}

func (s *SandboxService) waitForContainerStopped(ctx context.Context, sandboxId string, timeout time.Duration) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return status.Errorf(codes.DeadlineExceeded, "timeout waiting for container %s to stop", sandboxId)
		case <-ticker.C:
			c, err := s.dockerClient.ContainerInspect(ctx, sandboxId)
			if err != nil {
				return common.MapDockerError(err)
			}

			if !c.State.Running {
				return nil
			}
		}
	}
}
