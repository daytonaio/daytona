// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
)

func (s *TargetService) GetTarget(ctx context.Context, filter *stores.TargetFilter, params services.TargetRetrievalParams) (*services.TargetDTO, error) {
	tg, err := s.targetStore.Find(ctx, filter)
	if err != nil {
		return nil, stores.ErrTargetNotFound
	}

	state := tg.GetState()

	if state.Name == models.ResourceStateNameDeleted && !params.ShowDeleted {
		return nil, services.ErrTargetDeleted
	}

	var updatedWorkspaces []models.Workspace
	for _, w := range tg.Workspaces {
		wsState := w.GetState()
		if wsState.Name != models.ResourceStateNameDeleted {
			updatedWorkspaces = append(updatedWorkspaces, w)
		}
	}
	tg.Workspaces = updatedWorkspaces

	return &services.TargetDTO{
		Target: *tg,
		State:  state,
	}, nil
}
