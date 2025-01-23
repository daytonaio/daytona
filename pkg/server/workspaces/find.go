// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
)

func (s *WorkspaceService) Find(ctx context.Context, workspaceId string, params services.WorkspaceRetrievalParams) (*services.WorkspaceDTO, error) {
	w, err := s.workspaceStore.Find(ctx, workspaceId)
	if err != nil {
		return nil, stores.ErrWorkspaceNotFound
	}

	state := w.GetState()

	if state.Name == models.ResourceStateNameDeleted && !params.ShowDeleted {
		return nil, services.ErrWorkspaceDeleted
	}

	return &services.WorkspaceDTO{
		Workspace: *w,
		State:     state,
	}, nil
}
