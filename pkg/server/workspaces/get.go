// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	log "github.com/sirupsen/logrus"
)

func (s *WorkspaceService) GetWorkspace(ctx context.Context, workspaceId string, params services.WorkspaceRetrievalParams) (*services.WorkspaceDTO, error) {
	ws, err := s.workspaceStore.Find(workspaceId)
	if err != nil {
		return nil, stores.ErrWorkspaceNotFound
	}

	state := ws.GetState()

	if state.Name == models.ResourceStateNameDeleted && !params.ShowDeleted {
		return nil, services.ErrWorkspaceDeleted
	}

	response := &services.WorkspaceDTO{
		Workspace: *ws,
		State:     state,
	}

	if !params.Verbose {
		return response, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resultCh := make(chan provisioner.WorkspaceInfoResult, 1)

	go func() {
		workspaceInfo, err := s.provisioner.GetWorkspaceInfo(ctx, ws)
		resultCh <- provisioner.WorkspaceInfoResult{Info: workspaceInfo, Err: err}
	}()

	select {
	case res := <-resultCh:
		if res.Err != nil {
			log.Error(fmt.Errorf("failed to get workspace info for %s: %v", ws.Name, res.Err))
			return nil, res.Err
		}

		response.Info = res.Info
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			log.Warn(fmt.Sprintf("timeout getting workspace info for %s", ws.Name))
		} else {
			log.Warn(fmt.Sprintf("cancelled getting workspace info for %s", ws.Name))
		}
	}

	return response, nil
}
