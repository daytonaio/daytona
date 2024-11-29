// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server/workspacetemplates/dto"
	"github.com/daytonaio/daytona/pkg/stores"
)

type IWorkspaceTemplateService interface {
	Save(workspaceTemplate *models.WorkspaceTemplate) error
	Find(filter *stores.WorkspaceTemplateFilter) (*models.WorkspaceTemplate, error)
	List(filter *stores.WorkspaceTemplateFilter) ([]*models.WorkspaceTemplate, error)
	SetDefault(workspaceTemplateName string) error
	Delete(workspaceTemplateName string, force bool) []error

	SetPrebuild(workspaceTemplateName string, createPrebuildDto dto.CreatePrebuildDTO) (*dto.PrebuildDTO, error)
	FindPrebuild(workspaceTemplateFilter *stores.WorkspaceTemplateFilter, prebuildFilter *stores.PrebuildFilter) (*dto.PrebuildDTO, error)
	ListPrebuilds(workspaceTemplateFilter *stores.WorkspaceTemplateFilter, prebuildFilter *stores.PrebuildFilter) ([]*dto.PrebuildDTO, error)
	DeletePrebuild(workspaceTemplateName string, id string, force bool) []error

	StartRetentionPoller() error
	EnforceRetentionPolicy() error
	ProcessGitEvent(gitprovider.GitEventData) error
}
