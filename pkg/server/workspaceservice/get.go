// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaceservice

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/server/api/controllers/workspace/dto"
	"github.com/daytonaio/daytona/pkg/server/db"
	"github.com/daytonaio/daytona/pkg/server/targets"
)

func GetWorkspace(workspaceId string) (*dto.WorkspaceDTO, error) {
	workspace, err := db.FindWorkspaceByIdOrName(workspaceId)
	if err != nil {
		return nil, errors.New("workspace not found")
	}

	target, err := targets.GetTarget(workspace.Target)
	if err != nil {
		return nil, err
	}

	workspaceInfo, err := provisioner.GetWorkspaceInfo(workspace, target)
	if err != nil {
		return nil, err
	}

	response := dto.WorkspaceDTO{
		Workspace: *workspace,
		Info:      workspaceInfo,
	}

	return &response, nil
}
