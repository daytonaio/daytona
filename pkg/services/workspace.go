// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"context"
	"io"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/models"
)

type IWorkspaceService interface {
	CreateWorkspace(ctx context.Context, req CreateWorkspaceDTO) (*models.Workspace, error)
	GetWorkspace(ctx context.Context, workspaceId string, verbose bool) (*WorkspaceDTO, error)
	ListWorkspaces(ctx context.Context, verbose bool) ([]WorkspaceDTO, error)
	StartWorkspace(ctx context.Context, workspaceId string) error
	StopWorkspace(ctx context.Context, workspaceId string) error
	RemoveWorkspace(ctx context.Context, workspaceId string) error
	ForceRemoveWorkspace(ctx context.Context, workspaceId string) error

	GetWorkspaceLogReader(workspaceId string) (io.Reader, error)
	SetWorkspaceState(workspaceId string, state *models.WorkspaceState) (*models.Workspace, error)
}

type WorkspaceDTO struct {
	models.Workspace
	Info *models.WorkspaceInfo `json:"info" validate:"optional"`
} //	@name	WorkspaceDTO

type CreateWorkspaceDTO struct {
	Id                  string                   `json:"id" validate:"required"`
	Name                string                   `json:"name" validate:"required"`
	Image               *string                  `json:"image,omitempty" validate:"optional"`
	User                *string                  `json:"user,omitempty" validate:"optional"`
	BuildConfig         *models.BuildConfig      `json:"buildConfig,omitempty" validate:"optional"`
	Source              CreateWorkspaceSourceDTO `json:"source" validate:"required"`
	EnvVars             map[string]string        `json:"envVars" validate:"required"`
	TargetId            string                   `json:"targetId" validate:"required"`
	GitProviderConfigId *string                  `json:"gitProviderConfigId" validate:"optional"`
} //	@name	CreateWorkspaceDTO

func (c *CreateWorkspaceDTO) ToWorkspace() *models.Workspace {
	w := &models.Workspace{
		Id:                  c.Id,
		Name:                c.Name,
		BuildConfig:         c.BuildConfig,
		Repository:          c.Source.Repository,
		EnvVars:             c.EnvVars,
		TargetId:            c.TargetId,
		GitProviderConfigId: c.GitProviderConfigId,
	}

	if c.Image != nil {
		w.Image = *c.Image
	}

	if c.User != nil {
		w.User = *c.User
	}

	return w
}

type CreateWorkspaceSourceDTO struct {
	Repository *gitprovider.GitRepository `json:"repository" validate:"required"`
} // @name CreateWorkspaceSourceDTO
