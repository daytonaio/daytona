// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
)

func (s *TargetService) List(ctx context.Context, filter *stores.TargetFilter, params services.TargetRetrievalParams) ([]services.TargetDTO, error) {
	targets, err := s.targetStore.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	response := []services.TargetDTO{}

	for _, t := range targets {
		state := t.GetState()

		if state.Name == models.ResourceStateNameDeleted && !params.ShowDeleted {
			continue
		}

		var updatedWorkspaces []models.Workspace
		for _, w := range t.Workspaces {
			wsState := w.GetState()
			if wsState.Name != models.ResourceStateNameDeleted {
				updatedWorkspaces = append(updatedWorkspaces, w)
			}
		}
		t.Workspaces = updatedWorkspaces

		response = append(response, services.TargetDTO{
			Target: *t,
			State:  state,
		})
	}

	return response, nil
}
