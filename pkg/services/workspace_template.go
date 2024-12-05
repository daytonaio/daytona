// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"context"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type IWorkspaceTemplateService interface {
	Save(ctx context.Context, workspaceTemplate *models.WorkspaceTemplate) error
	Find(ctx context.Context, filter *stores.WorkspaceTemplateFilter) (*models.WorkspaceTemplate, error)
	List(ctx context.Context, filter *stores.WorkspaceTemplateFilter) ([]*models.WorkspaceTemplate, error)
	SetDefault(ctx context.Context, workspaceTemplateName string) error
	Delete(ctx context.Context, workspaceTemplateName string, force bool) []error

	SetPrebuild(ctx context.Context, workspaceTemplateName string, createPrebuildDto CreatePrebuildDTO) (*PrebuildDTO, error)
	FindPrebuild(ctx context.Context, workspaceTemplateFilter *stores.WorkspaceTemplateFilter, prebuildFilter *stores.PrebuildFilter) (*PrebuildDTO, error)
	ListPrebuilds(ctx context.Context, workspaceTemplateFilter *stores.WorkspaceTemplateFilter, prebuildFilter *stores.PrebuildFilter) ([]*PrebuildDTO, error)
	DeletePrebuild(ctx context.Context, workspaceTemplateName string, id string, force bool) []error

	StartRetentionPoller(ctx context.Context) error
	EnforceRetentionPolicy(ctx context.Context) error
	ProcessGitEvent(ctx context.Context, gitEventData gitprovider.GitEventData) error
}

type CreateWorkspaceTemplateDTO struct {
	Name                string              `json:"name" validate:"required"`
	Image               *string             `json:"image,omitempty" validate:"optional"`
	User                *string             `json:"user,omitempty" validate:"optional"`
	BuildConfig         *models.BuildConfig `json:"buildConfig,omitempty" validate:"optional"`
	RepositoryUrl       string              `json:"repositoryUrl" validate:"required"`
	EnvVars             map[string]string   `json:"envVars" validate:"required"`
	GitProviderConfigId *string             `json:"gitProviderConfigId" validate:"optional"`
} // @name CreateWorkspaceTemplateDTO

type PrebuildDTO struct {
	Id                    string   `json:"id" validate:"required"`
	WorkspaceTemplateName string   `json:"workspaceTemplateName" validate:"required"`
	Branch                string   `json:"branch" validate:"required"`
	CommitInterval        *int     `json:"commitInterval" validate:"optional"`
	TriggerFiles          []string `json:"triggerFiles" validate:"optional"`
	Retention             int      `json:"retention" validate:"required"`
} // @name PrebuildDTO

type CreatePrebuildDTO struct {
	Id             *string  `json:"id" validate:"optional"`
	Branch         string   `json:"branch" validate:"optional"`
	CommitInterval *int     `json:"commitInterval" validate:"optional"`
	TriggerFiles   []string `json:"triggerFiles" validate:"optional"`
	Retention      int      `json:"retention" validate:"required"`
} // @name CreatePrebuildDTO
