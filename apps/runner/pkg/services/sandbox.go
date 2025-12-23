// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package services

import (
	"context"
	"fmt"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/runner/pkg/cache"
	"github.com/daytonaio/runner/pkg/docker"
	"github.com/daytonaio/runner/pkg/models"
	"github.com/daytonaio/runner/pkg/models/enums"

	log "github.com/sirupsen/logrus"
)

type SandboxService struct {
	backupInfoCache *cache.BackupInfoCache
	docker          *docker.DockerClient
}

func NewSandboxService(backupInfoCache *cache.BackupInfoCache, docker *docker.DockerClient) *SandboxService {
	return &SandboxService{
		backupInfoCache: backupInfoCache,
		docker:          docker,
	}
}

func (s *SandboxService) GetSandboxInfo(ctx context.Context, sandboxId string) (*models.SandboxInfo, error) {
	sandboxState, err := s.docker.GetSandboxState(ctx, sandboxId)
	if err != nil {
		return &models.SandboxInfo{
			SandboxState:      enums.SandboxStateUnknown,
			BackupState:       enums.BackupStateNone,
			BackupErrorReason: nil,
		}, err
	}

	if sandboxState == enums.SandboxStateDestroyed {
		log.Warnf("Sandbox returned DESTROYED without error for sandbox %s indicating that sandbox is in DEAD state", sandboxId)

		err := s.backupInfoCache.Delete(ctx, sandboxId)
		if err != nil {
			log.Warnf("Failed to delete backup info cache for destroyed sandbox %s: %v", sandboxId, err)
		}

		return &models.SandboxInfo{
			SandboxState:      enums.SandboxStateUnknown,
			BackupState:       enums.BackupStateNone,
			BackupErrorReason: nil,
		}, common_errors.NewNotFoundError(fmt.Errorf("sandbox %s not found", sandboxId))
	}

	backupInfo, err := s.backupInfoCache.Get(ctx, sandboxId)
	if err != nil {
		errReason := err.Error()
		return &models.SandboxInfo{
			SandboxState:      sandboxState,
			BackupState:       enums.BackupStateNone,
			BackupErrorReason: &errReason,
		}, nil
	}

	return &models.SandboxInfo{
		SandboxState:      sandboxState,
		BackupState:       backupInfo.State,
		BackupErrorReason: backupInfo.ErrReason,
	}, nil
}
