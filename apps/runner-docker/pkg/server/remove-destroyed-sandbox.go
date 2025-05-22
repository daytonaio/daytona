// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package server

import (
	"context"
	"fmt"

	"github.com/daytonaio/runner-docker/pkg/common"
	"github.com/daytonaio/runner-docker/pkg/models/enums"
	pb "github.com/daytonaio/runner/proto"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/errdefs"
	log "github.com/sirupsen/logrus"
)

func (s *RunnerServer) RemoveDestroyedSandbox(ctx context.Context, req *pb.RemoveDestroyedSandboxRequest) (*pb.RemoveDestroyedSandboxResponse, error) {
	// Check if container exists and is in destroyed state
	state, err := s.getSandboxState(ctx, req.SandboxId)
	if err != nil {
		return nil, err
	}

	if state != enums.SandboxStateDestroyed {
		return nil, common.NewBadRequestError(fmt.Errorf("sandbox %s is not in destroyed state", req.SandboxId))
	}

	// Remove the container
	err = s.dockerClient.ContainerRemove(ctx, req.SandboxId, container.RemoveOptions{
		Force: true,
	})
	if err != nil {
		if errdefs.IsNotFound(err) {
			return &pb.RemoveDestroyedSandboxResponse{
				Message: fmt.Sprintf("Destroyed sandbox %s already removed", req.SandboxId),
			}, nil
		}
		return nil, err
	}

	log.Infof("Destroyed sandbox %s removed successfully", req.SandboxId)

	return &pb.RemoveDestroyedSandboxResponse{
		Message: fmt.Sprintf("Destroyed sandbox %s removed", req.SandboxId),
	}, nil
}
