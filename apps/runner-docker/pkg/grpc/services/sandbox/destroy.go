// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"fmt"
	"time"

	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
	"github.com/daytonaio/runner-docker/internal/metrics"
	"github.com/daytonaio/runner-docker/pkg/common"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/errdefs"
)

func (s *SandboxService) DestroySandbox(ctx context.Context, req *pb.DestroySandboxRequest) (*pb.DestroySandboxResponse, error) {
	startTime := time.Now()
	defer func() {
		obs, err := metrics.ContainerOperationDuration.GetMetricWithLabelValues("destroy")
		if err == nil {
			obs.Observe(time.Since(startTime).Seconds())
		}
	}()

	state, _ := s.getSandboxState(ctx, req.GetSandboxId())
	if state == pb.SandboxState_SANDBOX_STATE_DESTROYED || state == pb.SandboxState_SANDBOX_STATE_DESTROYING {
		metrics.SuccessCounterInc(metrics.DestroySandboxOperation)
		return &pb.DestroySandboxResponse{
			Message: fmt.Sprintf("Sandbox %s already destroyed", req.GetSandboxId()),
		}, nil
	}

	s.cache.SetSandboxState(ctx, req.GetSandboxId(), pb.SandboxState_SANDBOX_STATE_DESTROYING)

	_, err := s.dockerClient.ContainerInspect(ctx, req.GetSandboxId())
	if err != nil {
		if errdefs.IsNotFound(err) {
			s.cache.SetSandboxState(ctx, req.GetSandboxId(), pb.SandboxState_SANDBOX_STATE_DESTROYED)
			metrics.SuccessCounterInc(metrics.DestroySandboxOperation)

			return &pb.DestroySandboxResponse{
				Message: fmt.Sprintf("Sandbox %s already destroyed", req.GetSandboxId()),
			}, nil
		}

		metrics.FailureCounterInc(metrics.DestroySandboxOperation)
		return nil, common.MapDockerError(err)
	}

	err = s.dockerClient.ContainerRemove(ctx, req.GetSandboxId(), container.RemoveOptions{
		Force: true,
	})
	if err != nil {
		if errdefs.IsNotFound(err) {
			s.cache.SetSandboxState(ctx, req.GetSandboxId(), pb.SandboxState_SANDBOX_STATE_DESTROYED)
			metrics.SuccessCounterInc(metrics.DestroySandboxOperation)

			return &pb.DestroySandboxResponse{
				Message: fmt.Sprintf("Sandbox %s already destroyed", req.GetSandboxId()),
			}, nil
		}

		metrics.FailureCounterInc(metrics.DestroySandboxOperation)
		return nil, common.MapDockerError(err)
	}

	s.cache.SetSandboxState(ctx, req.GetSandboxId(), pb.SandboxState_SANDBOX_STATE_DESTROYED)

	metrics.SuccessCounterInc(metrics.DestroySandboxOperation)

	return &pb.DestroySandboxResponse{
		Message: fmt.Sprintf("Sandbox %s destroyed", req.GetSandboxId()),
	}, nil
}
