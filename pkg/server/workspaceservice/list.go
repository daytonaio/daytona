// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaceservice

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/server/api/controllers/workspace/dto"
	"github.com/daytonaio/daytona/pkg/server/db"
	"github.com/daytonaio/daytona/pkg/server/targets"
	"github.com/daytonaio/daytona/pkg/types"
	log "github.com/sirupsen/logrus"
)

func ListWorkspaces(verbose bool) ([]dto.WorkspaceDTO, error) {
	workspaces, err := db.ListWorkspaces()

	if err != nil {
		return nil, err
	}

	response := []dto.WorkspaceDTO{}

	for _, workspace := range workspaces {
		var workspaceInfo *types.WorkspaceInfo
		if verbose {
			target, err := targets.GetTarget(workspace.Target)
			if err != nil {
				log.Error(fmt.Errorf("failed to get target for %s", workspace.Target))
				continue
			}

			workspaceInfo, err = provisioner.GetWorkspaceInfo(workspace, target)
			if err != nil {
				log.Error(fmt.Errorf("failed to get workspace info for %s", workspace.Name))
				// return
			}
		}

		response = append(response, dto.WorkspaceDTO{
			Workspace: *workspace,
			Info:      workspaceInfo,
		})
	}

	return response, nil
}
