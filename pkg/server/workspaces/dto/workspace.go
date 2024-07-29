// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	projectconfig_dto "github.com/daytonaio/daytona/pkg/server/projectconfig/dto"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/daytonaio/daytona/pkg/workspace/project"
)

type WorkspaceDTO struct {
	workspace.Workspace
	Info *workspace.WorkspaceInfo
} //	@name	WorkspaceDTO

type ProjectDTO struct {
	project.Project
	Info *project.ProjectInfo
} //	@name	ProjectDTO

type CreateWorkspaceDTO struct {
	Id       string                                     `json:"id"`
	Name     string                                     `json:"name"`
	Target   string                                     `json:"target"`
	Projects []projectconfig_dto.CreateProjectConfigDTO `json:"projects" validate:"required,gt=0,dive"`
} //	@name	CreateWorkspaceDTO
