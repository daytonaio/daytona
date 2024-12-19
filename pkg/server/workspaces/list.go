// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
)

func (s *WorkspaceService) ListWorkspaces(ctx context.Context, params services.WorkspaceRetrievalParams) ([]services.WorkspaceDTO, error) {
	workspaces, err := s.workspaceStore.List(ctx)
	if err != nil {
		return nil, err
	}

	response := []services.WorkspaceDTO{}

	for _, ws := range workspaces {
		state := ws.GetState()

		if state.Name == models.ResourceStateNameDeleted && !params.ShowDeleted {
			continue
		}

		response = append(response, services.WorkspaceDTO{
			Workspace: *ws,
			State:     state,
		})
	}

	return response, nil
}
