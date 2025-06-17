// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"fmt"
	"io"

	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
	"github.com/daytonaio/runner-docker/internal/util"
	"github.com/daytonaio/runner-docker/pkg/common"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/errdefs"
	"github.com/docker/docker/pkg/jsonmessage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *SandboxService) CreateBackup(ctx context.Context, req *pb.CreateBackupRequest) (*pb.CreateBackupResponse, error) {
	s.cache.SetBackupState(ctx, req.GetSandboxId(), pb.BackupState_BACKUP_STATE_PENDING)

	s.log.Info("Creating snapshot for container", "sandboxId", req.GetSandboxId())

	s.cache.SetBackupState(ctx, req.GetSandboxId(), pb.BackupState_BACKUP_STATE_IN_PROGRESS)

	err := s.commitContainer(ctx, req.GetSandboxId(), req.GetSnapshot())
	if err != nil {
		return nil, err
	}

	if s.logWriter != nil {
		s.logWriter.Write([]byte(fmt.Sprintf("Pushing snapshot %s...", req.GetSnapshot())))
	}

	responseBody, err := s.dockerClient.ImagePush(ctx, req.GetSnapshot(), image.PushOptions{
		RegistryAuth: common.GetRegistryAuth(req.GetRegistry()),
	})
	if err != nil {
		return nil, common.MapDockerError(err)
	}
	defer responseBody.Close()

	err = jsonmessage.DisplayJSONMessagesStream(responseBody, io.Writer(&util.DebugLogWriter{}), 0, true, nil)
	if err != nil {
		return nil, err
	}

	if s.logWriter != nil {
		s.logWriter.Write([]byte(fmt.Sprintf("Snapshot %s pushed successfully", req.GetSnapshot())))
	}

	s.cache.SetBackupState(ctx, req.GetSandboxId(), pb.BackupState_BACKUP_STATE_COMPLETED)

	s.log.Info("Backup created successfully", "snapshot", req.GetSnapshot(), "sandboxId", req.GetSandboxId())

	// TODO: Duplicate code from RemoveSnapshot, check later
	_, err = s.dockerClient.ImageRemove(ctx, req.GetSnapshot(), image.RemoveOptions{
		Force:         true,
		PruneChildren: true,
	})
	if err != nil {
		if errdefs.IsNotFound(err) {
			s.log.Info("Snapshot already removed and not found", "snapshot", req.GetSnapshot())
		}
		s.log.Error("Error removing snapshot", "snapshot", req.GetSnapshot(), "error", err)
	} else {
		s.log.Info("Snapshot removed successfully", "snapshot", req.GetSnapshot())
	}

	return &pb.CreateBackupResponse{
		Message: fmt.Sprintf("Backup created for sandbox %s", req.GetSandboxId()),
	}, nil
}

func (s *SandboxService) commitContainer(ctx context.Context, sandboxId, snapshot string) error {
	const maxRetries = 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		s.log.Info("Committing container", "sandboxId", sandboxId, "attempt", attempt, "maxRetries", maxRetries)

		commitResp, err := s.dockerClient.ContainerCommit(ctx, sandboxId, container.CommitOptions{
			Reference: snapshot,
			Pause:     false,
		})
		if err == nil {
			s.log.Info("Container committed successfully", "sandboxId", sandboxId, "snapshotId", commitResp.ID)
			return nil
		}

		if attempt < maxRetries {
			s.log.Warn("Failed to commit container", "sandboxId", sandboxId, "attempt", attempt, "maxRetries", maxRetries, "error", err)
			continue
		}

		return status.Error(codes.Internal, fmt.Errorf("failed to commit container after %d attempts: %w", maxRetries, err).Error())
	}

	return nil
}
