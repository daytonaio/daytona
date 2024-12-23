// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

func (s *WorkspaceService) SetWorkspaceMetadata(ctx context.Context, workspaceId string, metadata *models.WorkspaceMetadata) (*models.WorkspaceMetadata, error) {
	m, err := s.workspaceMetadataStore.Find(ctx, workspaceId)
	if err != nil {
		return nil, stores.ErrWorkspaceMetadataNotFound
	}

	m.GitStatus = metadata.GitStatus
	m.Uptime = metadata.Uptime
	m.UpdatedAt = metadata.UpdatedAt
	return m, s.workspaceMetadataStore.Save(ctx, m)
}
