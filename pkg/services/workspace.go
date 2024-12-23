// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"context"
	"errors"
	"io"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/models"
)

type IWorkspaceService interface {
	CreateWorkspace(ctx context.Context, req CreateWorkspaceDTO) (*WorkspaceDTO, error)
	GetWorkspace(ctx context.Context, workspaceId string, params WorkspaceRetrievalParams) (*WorkspaceDTO, error)
	ListWorkspaces(ctx context.Context, params WorkspaceRetrievalParams) ([]WorkspaceDTO, error)
	StartWorkspace(ctx context.Context, workspaceId string) error
	StopWorkspace(ctx context.Context, workspaceId string) error
	RemoveWorkspace(ctx context.Context, workspaceId string) error
	ForceRemoveWorkspace(ctx context.Context, workspaceId string) error
	UpdateWorkspaceProviderMetadata(ctx context.Context, workspaceId, metadata string) error

	GetWorkspaceLogReader(ctx context.Context, workspaceId string) (io.Reader, error)
	GetWorkspaceLogWriter(ctx context.Context, workspaceId string) (io.WriteCloser, error)
	SetWorkspaceMetadata(ctx context.Context, workspaceId string, metadata *models.WorkspaceMetadata) (*models.WorkspaceMetadata, error)
}

type WorkspaceDTO struct {
	models.Workspace
	State models.ResourceState `json:"state" validate:"required"`
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
	GitProviderConfigId *string                  `json:"gitProviderConfigId,omitempty" validate:"optional"`
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

type WorkspaceRetrievalParams struct {
	ShowDeleted bool
}

var (
	ErrWorkspaceAlreadyExists   = errors.New("workspace already exists")
	ErrWorkspaceDeleted         = errors.New("workspace is deleted")
	ErrInvalidWorkspaceName     = errors.New("workspace name is not valid. Only [a-zA-Z0-9-_.] are allowed")
	ErrInvalidWorkspaceTemplate = errors.New("workspace template is invalid")
)

func IsWorkspaceDeleted(err error) bool {
	return errors.Is(err, ErrWorkspaceDeleted)
}
