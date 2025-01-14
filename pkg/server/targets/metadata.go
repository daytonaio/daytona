// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

func (s *TargetService) UpdateMetadata(ctx context.Context, targetId string, metadata *models.TargetMetadata) (*models.TargetMetadata, error) {
	m, err := s.targetMetadataStore.Find(ctx, targetId)
	if err != nil {
		return nil, stores.ErrTargetMetadataNotFound
	}

	m.Uptime = metadata.Uptime
	m.UpdatedAt = metadata.UpdatedAt
	return m, s.targetMetadataStore.Save(ctx, m)
}
