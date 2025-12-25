// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package services

import (
	"context"

	"github.com/daytonaio/runner-win/pkg/cache"
	"github.com/daytonaio/runner-win/pkg/libvirt"
	"github.com/daytonaio/runner-win/pkg/models"
	"github.com/daytonaio/runner-win/pkg/models/enums"

	log "github.com/sirupsen/logrus"
)

type SandboxService struct {
	statesCache *cache.StatesCache
	libvirt     *libvirt.LibVirt
}

func NewSandboxService(statesCache *cache.StatesCache, libvirt *libvirt.LibVirt) *SandboxService {
	return &SandboxService{
		statesCache: statesCache,
		libvirt:     libvirt,
	}
}

func (s *SandboxService) GetSandboxStatesInfo(ctx context.Context, sandboxId string) *models.CachedStates {
	sandboxState, err := s.libvirt.DeduceSandboxState(ctx, sandboxId)
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
		err := s.libvirt.Destroy(ctx, sandboxId)
		if err != nil {
			return err
		}
	}

	return nil
}
