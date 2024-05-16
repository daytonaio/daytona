// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

func (s *WorkspaceService) RemoveWorkspace(workspaceId string) error {
	workspace, err := s.WorkspaceStore.Find(workspaceId)
	if err != nil {
		return ErrWorkspaceNotFound
	}

	log.Infof("Destroying workspace %s", workspace.Id)

	target, err := s.TargetStore.Find(workspace.Target)
	if err != nil {
		return err
	}

	for _, project := range workspace.Projects {
		//	todo: go routines
		err := s.Provisioner.DestroyProject(project, target)
		if err != nil {
			return err
		}
	}

	err = s.Provisioner.DestroyWorkspace(workspace, target)
	if err != nil {
		return err
	}

	// Should not fail the whole operation if the API key cannot be revoked
	err = s.ApiKeyService.Revoke(workspace.Id)
	if err != nil {
		log.Error(err)
	}

	for _, project := range workspace.Projects {
		err := s.ApiKeyService.Revoke(fmt.Sprintf("%s/%s", workspace.Id, project.Name))
		if err != nil {
			// Should not fail the whole operation if the API key cannot be revoked
			log.Error(err)
		}
		projectLogger := s.LoggerFactory.CreateProjectLogger(workspace.Id, project.Name)
		err = projectLogger.Cleanup()
		if err != nil {
			// Should not fail the whole operation if the project logger cannot be cleaned up
			log.Error(err)
		}
	}

	logger := s.LoggerFactory.CreateWorkspaceLogger(workspace.Id)
	err = logger.Cleanup()
	if err != nil {
		// Should not fail the whole operation if the workspace logger cannot be cleaned up
		log.Error(err)
	}

	err = s.WorkspaceStore.Delete(workspace)
	if err != nil {
		return err
	}

	log.Infof("Workspace %s destroyed", workspace.Id)
	return nil
}
