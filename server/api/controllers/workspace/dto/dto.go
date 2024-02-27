package dto

import "github.com/daytonaio/daytona/common/types"

type CreateWorkspace struct {
	Name         string
	Repositories []string
	Provisioner  string
} //	@name	CreateWorkspace

type Workspace struct {
	types.Workspace
	Info *types.WorkspaceInfo
} //	@name	Workspace

type Project struct {
	types.Project
	Info *types.ProjectInfo
} //	@name	Project
