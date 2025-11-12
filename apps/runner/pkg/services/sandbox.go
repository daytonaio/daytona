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
)

type SandboxService struct {
	statesCache *cache.StatesCache
	docker      *docker.DockerClient
}

func NewSandboxService(statesCache *cache.StatesCache, docker *docker.DockerClient) *SandboxService {
	return &SandboxService{
		statesCache: statesCache,
		docker:      docker,
	}
}

func (s *SandboxService) GetSandboxStatesInfo(ctx context.Context, sandboxId string) (*models.CachedStates, error) {
	sandboxState, err := s.docker.GetSandboxState(ctx, sandboxId)
	if err == nil {
		s.statesCache.SetSandboxState(ctx, sandboxId, sandboxState)
	}

	if err != nil && (common_errors.IsNotFoundError(err) || sandboxState == enums.SandboxStateDestroyed) {
		return &models.CachedStates{
			SandboxState:      enums.SandboxStateUnknown,
			BackupState:       enums.BackupStateNone,
			BackupErrorReason: nil,
		}, common_errors.NewNotFoundError(fmt.Errorf("sandbox %s not found", sandboxId))
	}

	data, err := s.statesCache.Get(ctx, sandboxId)
	if err != nil {
		return &models.CachedStates{
			SandboxState:      enums.SandboxStateUnknown,
			BackupState:       enums.BackupStateNone,
			BackupErrorReason: nil,
		}, err
	}

	return data, nil
}
