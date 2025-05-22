// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package server

import (
	"context"
	"fmt"
	"io"

	"github.com/daytonaio/runner-docker/internal/util"
	"github.com/daytonaio/runner-docker/pkg/models/enums"
	pb "github.com/daytonaio/runner/proto"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/pkg/jsonmessage"

	log "github.com/sirupsen/logrus"
)

func (s *RunnerServer) CreateSnapshot(ctx context.Context, req *pb.CreateSnapshotRequest) (*pb.CreateSnapshotResponse, error) {
	s.cache.SetSnapshotState(ctx, req.SandboxId, enums.SnapshotStatePending)

	log.Infof("Creating snapshot for container %s...", req.SandboxId)

	s.cache.SetSnapshotState(ctx, req.SandboxId, enums.SnapshotStateInProgress)

	err := s.commitContainer(ctx, req.SandboxId, req.Image)
	if err != nil {
		return nil, err
	}

	err = s.pushImage(ctx, req.Image, req.Registry)
	if err != nil {
		return nil, err
	}

	s.cache.SetSnapshotState(ctx, req.SandboxId, enums.SnapshotStateCompleted)

	log.Infof("Snapshot (%s) for container %s created successfully", req.Image, req.SandboxId)

	_, err = s.RemoveImage(ctx, &pb.RemoveImageRequest{
		Image: req.Image,
		Force: true,
	})
	if err != nil {
		log.Errorf("Error removing image %s: %v", req.Image, err)
	}

	return &pb.CreateSnapshotResponse{
		Message: fmt.Sprintf("Snapshot created for workspace %s", req.SandboxId),
	}, nil
}

func (s *RunnerServer) commitContainer(ctx context.Context, sandboxId, imageName string) error {
	const maxRetries = 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Infof("Committing container %s (attempt %d/%d)...", sandboxId, attempt, maxRetries)

		commitResp, err := s.dockerClient.ContainerCommit(ctx, sandboxId, container.CommitOptions{
			Reference: imageName,
			Pause:     false,
		})
		if err == nil {
			log.Infof("Container %s committed successfully with image ID: %s", sandboxId, commitResp.ID)
			return nil
		}

		if attempt < maxRetries {
			log.Warnf("Failed to commit container %s (attempt %d/%d): %v", sandboxId, attempt, maxRetries, err)
			continue
		}

		return fmt.Errorf("failed to commit container after %d attempts: %w", maxRetries, err)
	}

	return nil
}

func (s *RunnerServer) pushImage(ctx context.Context, imageName string, reg *pb.Registry) error {
	log.Infof("Pushing image %s...", imageName)

	responseBody, err := s.dockerClient.ImagePush(ctx, imageName, image.PushOptions{
		RegistryAuth: getRegistryAuth(reg),
	})
	if err != nil {
		return err
	}
	defer responseBody.Close()

	err = jsonmessage.DisplayJSONMessagesStream(responseBody, io.Writer(&util.DebugLogWriter{}), 0, true, nil)
	if err != nil {
		return err
	}

	log.Infof("Image %s pushed successfully", imageName)

	return nil
}
