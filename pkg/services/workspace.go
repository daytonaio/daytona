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
	List(ctx context.Context, params WorkspaceRetrievalParams) ([]WorkspaceDTO, error)
	Find(ctx context.Context, workspaceId string, params WorkspaceRetrievalParams) (*WorkspaceDTO, error)
	Create(ctx context.Context, req CreateWorkspaceDTO) (*WorkspaceDTO, error)
	Start(ctx context.Context, workspaceId string) error
	Stop(ctx context.Context, workspaceId string) error
	Delete(ctx context.Context, workspaceId string) error
	ForceDelete(ctx context.Context, workspaceId string) error
	Restart(ctx context.Context, workspaceId string) error

	UpdateMetadata(ctx context.Context, workspaceId string, metadata *models.WorkspaceMetadata) (*models.WorkspaceMetadata, error)
	UpdateProviderMetadata(ctx context.Context, workspaceId, metadata string) error
	UpdateLastJob(ctx context.Context, workspaceId, jobId string) error
	UpdateLabels(ctx context.Context, workspaceId string, labels map[string]string) (*WorkspaceDTO, error)

	GetWorkspaceLogReader(ctx context.Context, workspaceId string) (io.Reader, error)
	GetWorkspaceLogWriter(ctx context.Context, workspaceId string) (io.WriteCloser, error)
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
	Labels              map[string]string        `json:"labels" validate:"required"`
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
		Labels:              c.Labels,
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
	Labels      map[string]string
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
