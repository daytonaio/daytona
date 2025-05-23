// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package services

import (
	"context"

	"github.com/daytonaio/runner/pkg/cache"
	"github.com/daytonaio/runner/pkg/docker"
	"github.com/daytonaio/runner/pkg/models"
	"github.com/daytonaio/runner/pkg/models/enums"
)

type SandboxService struct {
	cache  cache.IRunnerCache
	docker *docker.DockerClient
}

func NewSandboxService(cache cache.IRunnerCache, docker *docker.DockerClient) *SandboxService {
	return &SandboxService{
		cache:  cache,
		docker: docker,
	}
}

func (s *SandboxService) GetSandboxStatesInfo(ctx context.Context, sandboxId string) *models.CacheData {
	sandboxState, err := s.docker.DeduceSandboxState(ctx, sandboxId)
	if err == nil {
		s.cache.SetSandboxState(ctx, sandboxId, sandboxState)
	}

	data := s.cache.Get(ctx, sandboxId)

	if data == nil {
		return &models.CacheData{
			SandboxState:    enums.SandboxStateUnknown,
			BackupState:     enums.BackupStateNone,
			DestructionTime: nil,
		}
	}

	return data
}

func (s *SandboxService) RemoveDestroyedSandbox(ctx context.Context, sandboxId string) error {
	info := s.GetSandboxStatesInfo(ctx, sandboxId)

	if info != nil && info.SandboxState != enums.SandboxStateDestroyed && info.SandboxState != enums.SandboxStateDestroying {
		err := s.docker.Destroy(ctx, sandboxId)
		if err != nil {
			return err
		}
	}

	return nil
}
