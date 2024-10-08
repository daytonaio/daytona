// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/daytonaio/daytona/pkg/workspace/project"
	"github.com/daytonaio/daytona/pkg/workspace/project/buildconfig"
)

type WorkspaceDTO struct {
	workspace.Workspace
	Info *workspace.WorkspaceInfo `json:"info" validate:"optional"`
} //	@name	WorkspaceDTO

type ProjectDTO struct {
	project.Project
	Info *project.ProjectInfo `json:"info" validate:"optional"`
} //	@name	ProjectDTO

type CreateWorkspaceDTO struct {
	Id       string             `json:"id" validate:"required"`
	Name     string             `json:"name" validate:"required"`
	Target   string             `json:"target" validate:"required"`
	Projects []CreateProjectDTO `json:"projects" validate:"required,gt=0,dive"`
} //	@name	CreateWorkspaceDTO

type CreateProjectDTO struct {
	Name        string                   `json:"name" validate:"required"`
	Image       *string                  `json:"image,omitempty" validate:"optional"`
	User        *string                  `json:"user,omitempty" validate:"optional"`
	BuildConfig *buildconfig.BuildConfig `json:"buildConfig,omitempty" validate:"optional"`
	Source      CreateProjectSourceDTO   `json:"source" validate:"required"`
	EnvVars     map[string]string        `json:"envVars" validate:"required"`
} //	@name	CreateProjectDTO

type CreateProjectSourceDTO struct {
	Repository *gitprovider.GitRepository `json:"repository" validate:"required"`
} // @name CreateProjectSourceDTO
