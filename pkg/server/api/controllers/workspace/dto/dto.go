// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/workspace"
)

type CreateWorkspace struct {
	Name         string
	TargetId     string
	Repositories []gitprovider.GitRepository
} //	@name	CreateWorkspace

type WorkspaceDTO struct {
	workspace.Workspace
	Info *workspace.WorkspaceInfo
} //	@name	WorkspaceDTO

type ProjectDTO struct {
	workspace.Project
	Info *workspace.ProjectInfo
} //	@name	ProjectDTO
