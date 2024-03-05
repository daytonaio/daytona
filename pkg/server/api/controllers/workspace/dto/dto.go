package dto

import "github.com/daytonaio/daytona/pkg/types"

type CreateWorkspace struct {
	Name         string
	Target       string
	Repositories []types.Repository
} //	@name	CreateWorkspace

type Workspace struct {
	types.Workspace
	Info *types.WorkspaceInfo
} //	@name	Workspace

type Project struct {
	types.Project
	Info *types.ProjectInfo
} //	@name	Project
