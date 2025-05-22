// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package server

import (
	"context"
	"fmt"
	"time"

	"github.com/daytonaio/runner-docker/pkg/common"
	"github.com/daytonaio/runner-docker/pkg/models/enums"
	pb "github.com/daytonaio/runner/proto"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/errdefs"
)

func (s *RunnerServer) DestroySandbox(ctx context.Context, req *pb.DestroySandboxRequest) (*pb.DestroySandboxResponse, error) {
	startTime := time.Now()
	defer func() {
		obs, err := common.ContainerOperationDuration.GetMetricWithLabelValues("destroy")
		if err == nil {
			obs.Observe(time.Since(startTime).Seconds())
		}
	}()

	state, err := s.getSandboxState(ctx, req.SandboxId)
	if err != nil && state == enums.SandboxStateError {
		return nil, err
	}

	if state == enums.SandboxStateDestroyed || state == enums.SandboxStateDestroying {
		return &pb.DestroySandboxResponse{
			Message: fmt.Sprintf("Sandbox %s already destroyed", req.SandboxId),
		}, nil
	}

	s.cache.SetSandboxState(ctx, req.SandboxId, enums.SandboxStateDestroying)

	_, err = s.dockerClient.ContainerInspect(ctx, req.SandboxId)
	if err != nil {
		if errdefs.IsNotFound(err) {
			s.cache.SetSandboxState(ctx, req.SandboxId, enums.SandboxStateDestroyed)
		}
		return nil, err
	}

	err = s.dockerClient.ContainerRemove(ctx, req.SandboxId, container.RemoveOptions{
		Force: true,
	})
	if err != nil {
		if errdefs.IsNotFound(err) {
			s.cache.SetSandboxState(ctx, req.SandboxId, enums.SandboxStateDestroyed)
		}
		return nil, err
	}

	s.cache.SetSandboxState(ctx, req.SandboxId, enums.SandboxStateDestroyed)

	return &pb.DestroySandboxResponse{
		Message: fmt.Sprintf("Sandbox %s destroyed", req.SandboxId),
	}, nil
}
