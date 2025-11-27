// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package services

import (
	"context"

	"github.com/daytonaio/runner/pkg/cache"
	"github.com/daytonaio/runner/pkg/docker"
	"github.com/daytonaio/runner/pkg/models"
	"github.com/daytonaio/runner/pkg/models/enums"

	log "github.com/sirupsen/logrus"
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

func (s *SandboxService) GetSandboxStatesInfo(ctx context.Context, sandboxId string) *models.CachedStates {
	sandboxState, err := s.docker.DeduceSandboxState(ctx, sandboxId)
	if err != nil {
		log.Warnf("Failed to deduce sandbox %s state: %v", sandboxId, err)
	}

	s.statesCache.SetSandboxState(ctx, sandboxId, sandboxState)

	data, err := s.statesCache.Get(ctx, sandboxId)
	if err != nil {
		return &models.CachedStates{
			SandboxState:      enums.SandboxStateUnknown,
			BackupState:       enums.BackupStateNone,
			BackupErrorReason: nil,
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
