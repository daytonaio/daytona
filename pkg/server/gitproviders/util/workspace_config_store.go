// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package uitl

import (
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server/gitproviders"
	"github.com/daytonaio/daytona/pkg/server/workspaceconfigs"
)

type store struct {
	store workspaceconfigs.WorkspaceConfigStore
}

func (s *store) Save(workspaceConfig *models.WorkspaceConfig) error {
	return s.store.Save(workspaceConfig)
}

func (s *store) List(gitProviderConfigId string) ([]*models.WorkspaceConfig, error) {
	return s.store.List(&workspaceconfigs.WorkspaceConfigFilter{
		GitProviderConfigId: &gitProviderConfigId,
	})
}

func FromWorkspaceConfigStore(workspaceConfigStore workspaceconfigs.WorkspaceConfigStore) gitproviders.WorkspaceConfigStore {
	return &store{
		store: workspaceConfigStore,
	}
}
