// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"fmt"
	"time"

	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
	"github.com/daytonaio/runner-docker/internal/metrics"
	"github.com/daytonaio/runner-docker/pkg/services/common"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/errdefs"
	log "github.com/sirupsen/logrus"
)

func (s *SandboxService) DestroySandbox(ctx context.Context, req *pb.DestroySandboxRequest) (*pb.DestroySandboxResponse, error) {
	startTime := time.Now()
	defer func() {
		obs, err := metrics.ContainerOperationDuration.GetMetricWithLabelValues("destroy")
		if err == nil {
			obs.Observe(time.Since(startTime).Seconds())
		}
	}()

	state, err := s.getSandboxState(ctx, req.GetSandboxId())
	if err != nil && state == pb.SandboxState_SANDBOX_STATE_ERROR {
		return nil, err
	}

	if state == pb.SandboxState_SANDBOX_STATE_DESTROYED || state == pb.SandboxState_SANDBOX_STATE_DESTROYING {
		return &pb.DestroySandboxResponse{
			Message: fmt.Sprintf("Sandbox %s already destroyed", req.GetSandboxId()),
		}, nil
	}

	s.cache.SetSandboxState(ctx, req.GetSandboxId(), pb.SandboxState_SANDBOX_STATE_DESTROYING)

	_, err = s.dockerClient.ContainerInspect(ctx, req.GetSandboxId())
	if err != nil {
		if errdefs.IsNotFound(err) {
			s.cache.SetSandboxState(ctx, req.GetSandboxId(), pb.SandboxState_SANDBOX_STATE_DESTROYED)
		}
		return nil, common.MapDockerError(err)
	}

	err = s.dockerClient.ContainerRemove(ctx, req.GetSandboxId(), container.RemoveOptions{
		Force: true,
	})
	if err != nil {
		if errdefs.IsNotFound(err) {
			s.cache.SetSandboxState(ctx, req.GetSandboxId(), pb.SandboxState_SANDBOX_STATE_DESTROYED)
		}
		return nil, common.MapDockerError(err)
	}

	s.cache.SetSandboxState(ctx, req.GetSandboxId(), pb.SandboxState_SANDBOX_STATE_DESTROYED)

	return &pb.DestroySandboxResponse{
		Message: fmt.Sprintf("Sandbox %s destroyed", req.GetSandboxId()),
	}, nil
}

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
			return &pb.RemoveDestroyedSandboxResponse{
				Message: fmt.Sprintf("Destroyed sandbox %s already removed", req.GetSandboxId()),
			}, nil
		}
		return nil, common.MapDockerError(err)
	}

	log.Infof("Destroyed sandbox %s removed successfully", req.GetSandboxId())

	return &pb.RemoveDestroyedSandboxResponse{
		Message: fmt.Sprintf("Destroyed sandbox %s removed", req.GetSandboxId()),
	}, nil
}
