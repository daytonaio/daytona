// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/server/api/controllers/workspace/dto"
)

func (s *WorkspaceService) GetWorkspace(workspaceId string) (*dto.WorkspaceDTO, error) {
	workspace, err := s.workspaceStore.Find(workspaceId)
	if err != nil {
		return nil, errors.New("workspace not found")
	}

	providerName, targetName, err := s.parseTargetId(workspace.Target)
	if err != nil {
		return nil, err
	}

	target, err := s.targetStore.Find(providerName, targetName)
	if err != nil {
		return nil, err
	}

	workspaceInfo, err := s.provisioner.GetWorkspaceInfo(workspace, target)
	if err != nil {
		return nil, err
	}

	response := dto.WorkspaceDTO{
		Workspace: *workspace,
		Info:      workspaceInfo,
	}

	return &response, nil
}
