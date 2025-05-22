// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package server

import (
	"context"
	"fmt"
	"time"

	"github.com/daytonaio/runner-docker/pkg/models/enums"
	pb "github.com/daytonaio/runner/proto"
	"github.com/docker/docker/api/types/container"
)

func (s *RunnerServer) StopSandbox(ctx context.Context, req *pb.StopSandboxRequest) (*pb.StopSandboxResponse, error) {
	s.cache.SetSandboxState(ctx, req.SandboxId, enums.SandboxStateStopping)

	err := s.dockerClient.ContainerStop(ctx, req.SandboxId, container.StopOptions{
		Signal: "SIGKILL",
	})
	if err != nil {
		return nil, err
	}

	err = s.waitForContainerStopped(ctx, req.SandboxId, 10*time.Second)
	if err != nil {
		return nil, err
	}

	s.cache.SetSandboxState(ctx, req.SandboxId, enums.SandboxStateStopped)

	return &pb.StopSandboxResponse{
		Message: fmt.Sprintf("Sandbox %s stopped", req.SandboxId),
	}, nil
}

func (s *RunnerServer) waitForContainerStopped(ctx context.Context, sandboxId string, timeout time.Duration) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("timeout waiting for container %s to stop", sandboxId)
		case <-ticker.C:
			c, err := s.dockerClient.ContainerInspect(ctx, sandboxId)
			if err != nil {
				return err
			}

			if !c.State.Running {
				return nil
			}
		}
	}
}
