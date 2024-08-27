// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/server/workspaces/dto"
	"github.com/daytonaio/daytona/pkg/workspace"
	log "github.com/sirupsen/logrus"
)

func (s *WorkspaceService) ListWorkspaces(verbose bool) ([]dto.WorkspaceDTO, error) {
	workspaces, err := s.workspaceStore.List()
	if err != nil {
		return nil, err
	}

	response := []dto.WorkspaceDTO{}

	for _, w := range workspaces {
		var workspaceInfo *workspace.WorkspaceInfo
		if verbose {
			target, err := s.targetStore.Find(w.Target)
			if err != nil {
				log.Error(fmt.Errorf("failed to get target for %s", w.Target))
				continue
			}

			workspaceInfo, err = s.provisioner.GetWorkspaceInfo(w, target)
			if err != nil {
				log.Error(fmt.Errorf("failed to get workspace info for %s", w.Name))
			}
		}

		response = append(response, dto.WorkspaceDTO{
			Workspace: *w,
			Info:      workspaceInfo,
		})
	}

	return response, nil
}
