// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/workspace"
)

type WorkspaceDTO struct {
	workspace.Workspace
	Info *workspace.WorkspaceInfo
} //	@name	WorkspaceDTO

type ProjectDTO struct {
	workspace.Project
	Info *workspace.ProjectInfo
} //	@name	ProjectDTO

type CreateWorkspaceRequestProjectSource struct {
	Repository *gitprovider.GitRepository `json:"repository"`
} // @name CreateWorkspaceRequestProjectSource

type CreateWorkspaceRequestProject struct {
	Name              string                              `json:"name" validate:"required,gt=0"`
	Image             *string                             `json:"image,omitempty"`
	User              *string                             `json:"user,omitempty"`
	Build             *workspace.ProjectBuild             `json:"build,omitempty"`
	Source            CreateWorkspaceRequestProjectSource `json:"source"`
	EnvVars           map[string]string                   `json:"envVars"`
	PostStartCommands *[]string                           `json:"postStartCommands,omitempty"`
} // @name CreateWorkspaceRequestProject

type CreateWorkspaceRequest struct {
	Id       string                          `json:"id"`
	Name     string                          `json:"name"`
	Target   string                          `json:"target"`
	Projects []CreateWorkspaceRequestProject `json:"projects" validate:"required,gt=0,dive"`
} //	@name	CreateWorkspaceRequest
