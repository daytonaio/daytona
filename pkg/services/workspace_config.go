// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server/workspaceconfigs/dto"
	"github.com/daytonaio/daytona/pkg/stores"
)

type IWorkspaceConfigService interface {
	Save(workspaceConfig *models.WorkspaceConfig) error
	Find(filter *stores.WorkspaceConfigFilter) (*models.WorkspaceConfig, error)
	List(filter *stores.WorkspaceConfigFilter) ([]*models.WorkspaceConfig, error)
	SetDefault(workspaceConfigName string) error
	Delete(workspaceConfigName string, force bool) []error

	SetPrebuild(workspaceConfigName string, createPrebuildDto dto.CreatePrebuildDTO) (*dto.PrebuildDTO, error)
	FindPrebuild(workspaceConfigFilter *stores.WorkspaceConfigFilter, prebuildFilter *stores.PrebuildFilter) (*dto.PrebuildDTO, error)
	ListPrebuilds(workspaceConfigFilter *stores.WorkspaceConfigFilter, prebuildFilter *stores.PrebuildFilter) ([]*dto.PrebuildDTO, error)
	DeletePrebuild(workspaceConfigName string, id string, force bool) []error

	StartRetentionPoller() error
	EnforceRetentionPolicy() error
	ProcessGitEvent(gitprovider.GitEventData) error
}
