// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	log "github.com/sirupsen/logrus"
)

func (s *WorkspaceService) StopWorkspace(workspaceId string) error {
	workspace, err := s.workspaceStore.Find(workspaceId)
	if err != nil {
		return ErrWorkspaceNotFound
	}

	log.Info("Stopping workspace")

	target, err := s.targetStore.Find(workspace.Target)
	if err != nil {
		return err
	}

	for _, project := range workspace.Projects {
		//	todo: go routines
		err := s.provisioner.StopProject(project, target)
		if err != nil {
			return err
		}
	}

	err = s.provisioner.StopWorkspace(workspace, target)
	if err != nil {
		return err
	}

	log.Info("Workspace stopped")
	return nil
}

func (s *WorkspaceService) StopProject(workspaceId, projectName string) error {
	w, err := s.workspaceStore.Find(workspaceId)
	if err != nil {
		return ErrWorkspaceNotFound
	}

	project, err := w.GetProject(projectName)
	if err != nil {
		return ErrProjectNotFound
	}

	target, err := s.targetStore.Find(w.Target)
	if err != nil {
		return err
	}

	return s.provisioner.StopProject(project, target)
}
