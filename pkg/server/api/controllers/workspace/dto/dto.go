// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package dto

import "github.com/daytonaio/daytona/pkg/types"

type CreateWorkspace struct {
	Name         string
	Target       string
	Repositories []types.GitRepository
} //	@name	CreateWorkspace

type WorkspaceDTO struct {
	types.Workspace
	Info *types.WorkspaceInfo
} //	@name	WorkspaceDTO

type ProjectDTO struct {
	types.Project
	Info *types.ProjectInfo
} //	@name	ProjectDTO
