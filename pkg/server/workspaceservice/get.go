// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaceservice

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/server/api/controllers/workspace/dto"
	"github.com/daytonaio/daytona/pkg/server/db"
	"github.com/daytonaio/daytona/pkg/server/provisioner"
)

func GetWorkspace(workspaceId string) (*dto.WorkspaceDTO, error) {
	w, err := db.FindWorkspaceByIdOrName(workspaceId)
	if err != nil {
		return nil, errors.New("workspace not found")
	}

	workspaceInfo, err := provisioner.GetWorkspaceInfo(w)
	if err != nil {
		return nil, err
	}

	response := dto.WorkspaceDTO{
		Workspace: *w,
		Info:      workspaceInfo,
	}

	return &response, nil
}
