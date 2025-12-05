// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package services

import (
	"context"
	"fmt"
	"time"

	"github.com/daytonaio/apiclient"
	runnerapiclient "github.com/daytonaio/runner/pkg/apiclient"
	"github.com/daytonaio/runner/pkg/cache"
	log "github.com/sirupsen/logrus"

	"github.com/daytonaio/runner/pkg/models/enums"
)

type SandboxStateChangeService struct {
	apiClient   *apiclient.APIClient
	statesCache *cache.StatesCache
}

func NewSandboxStateChangeService(statesCache *cache.StatesCache) *SandboxStateChangeService {
	apiClient, err := runnerapiclient.GetApiClient()
	if err != nil {
		log.Errorf("failed to initialize API client: %v", err)
	}

	return &SandboxStateChangeService{
		statesCache: statesCache,
		apiClient:   apiClient,
	}
}

func (s *SandboxStateChangeService) OnStartEvent(ctx context.Context, sandboxId string) {
	err := s.handleStateChangeEvent(ctx, sandboxId, enums.SandboxStateStarted)
	if err != nil {
		log.Errorf("failed to handle state change event: %v", err)
	}
}

func (s *SandboxStateChangeService) OnStopEvent(ctx context.Context, sandboxId string) {
	err := s.handleStateChangeEvent(ctx, sandboxId, enums.SandboxStateStopped)
	if err != nil {
		log.Errorf("failed to handle state change event: %v", err)
	}
}

func (s *SandboxStateChangeService) handleStateChangeEvent(ctx context.Context, sandboxId string, localState enums.SandboxState) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	s.statesCache.SetSandboxState(ctxWithTimeout, sandboxId, localState)
	return s.sendStateChangeToApi(ctxWithTimeout, sandboxId, localState)
}

func (s *SandboxStateChangeService) sendStateChangeToApi(ctx context.Context, sandboxId string, localState enums.SandboxState) error {
	if s.apiClient == nil {
		return fmt.Errorf("API client is not initialized - cannot proceed with state change operation")
	}

	_, err := s.apiClient.SandboxAPI.UpdateSandboxState(ctx, sandboxId).UpdateSandboxStateDto(*apiclient.NewUpdateSandboxStateDto(
		string(s.convertToApiState(localState)),
	)).Execute()
	if err != nil {
		return fmt.Errorf("failed to update sandbox %s state to %s: %v", sandboxId, localState, err)
	}

	return nil
}

func (s *SandboxStateChangeService) convertToApiState(localState enums.SandboxState) apiclient.SandboxState {
	switch localState {
	case enums.SandboxStateCreating:
		return apiclient.SANDBOXSTATE_CREATING
	case enums.SandboxStateRestoring:
		return apiclient.SANDBOXSTATE_RESTORING
	case enums.SandboxStateDestroyed:
		return apiclient.SANDBOXSTATE_DESTROYED
	case enums.SandboxStateDestroying:
		return apiclient.SANDBOXSTATE_DESTROYING
	case enums.SandboxStateStarted:
		return apiclient.SANDBOXSTATE_STARTED
	case enums.SandboxStateStopped:
		return apiclient.SANDBOXSTATE_STOPPED
	case enums.SandboxStateStarting:
		return apiclient.SANDBOXSTATE_STARTING
	case enums.SandboxStateStopping:
		return apiclient.SANDBOXSTATE_STOPPING
	case enums.SandboxStateError:
		return apiclient.SANDBOXSTATE_ERROR
	case enums.SandboxStatePullingSnapshot:
		return apiclient.SANDBOXSTATE_PULLING_SNAPSHOT
	default:
		return apiclient.SANDBOXSTATE_UNKNOWN
	}
}
