// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package services

import (
	"context"
	"fmt"
	"log/slog"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/runner/pkg/cache"
	"github.com/daytonaio/runner/pkg/docker"
	"github.com/daytonaio/runner/pkg/models"
	"github.com/daytonaio/runner/pkg/models/enums"
)

type SandboxService struct {
	backupInfoCache *cache.BackupInfoCache
	docker          *docker.DockerClient
	log             *slog.Logger
}

func NewSandboxService(logger *slog.Logger, backupInfoCache *cache.BackupInfoCache, docker *docker.DockerClient) *SandboxService {
	return &SandboxService{
		log:             logger.With(slog.String("component", "sandbox_service")),
		backupInfoCache: backupInfoCache,
		docker:          docker,
	}
}

func (s *SandboxService) GetSandboxInfo(ctx context.Context, sandboxId string) (*models.SandboxInfo, error) {
	sandboxState, err := s.docker.GetSandboxState(ctx, sandboxId)
	if err != nil {
		s.log.Warn("Failed to get sandbox state", "sandboxId", sandboxId, "error", err)
		return nil, err
	}

	if sandboxState == enums.SandboxStateDestroyed {
		s.log.Warn("Sandbox returned sandbox state DESTROYED without an error, sandbox is in DEAD state", "sandbox_id", sandboxId)

		err := s.backupInfoCache.Delete(ctx, sandboxId)
		if err != nil {
			s.log.Warn("Failed to delete backup info cache for destroyed sandbox", "sandbox_id", sandboxId, "error", err)
		}

		return nil, common_errors.NewNotFoundError(fmt.Errorf("sandbox %s not found", sandboxId))
	}

	backupInfo, err := s.backupInfoCache.Get(ctx, sandboxId)
	if err != nil {
		return &models.SandboxInfo{
			SandboxState: sandboxState,
			BackupState:  enums.BackupStateNone,
		}, nil
	}

	sandboxInfo := &models.SandboxInfo{
		SandboxState: sandboxState,
		BackupState:  backupInfo.State,
	}

	var backupErrReason string
	if backupInfo.Error != nil {
		backupErrReason = backupInfo.Error.Error()
		sandboxInfo.BackupErrorReason = &backupErrReason
	}

	return sandboxInfo, nil
}
