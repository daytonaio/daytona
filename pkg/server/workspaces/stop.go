// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"errors"

	log "github.com/sirupsen/logrus"
)

func (s *WorkspaceService) StopWorkspace(workspaceId string) error {
	workspace, err := s.workspaceStore.Find(workspaceId)
	if err != nil {
		return errors.New("workspace not found")
	}

	log.Info("Stopping workspace")

	providerName, targetName, err := s.parseTargetId(workspace.Target)
	if err != nil {
		return err
	}

	target, err := s.targetStore.Find(providerName, targetName)
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
		return errors.New("workspace not found")
	}

	project, err := w.GetProject(projectName)
	if err != nil {
		return errors.New("project not found")
	}

	providerName, targetName, err := s.parseTargetId(w.Target)
	if err != nil {
		return err
	}

	target, err := s.targetStore.Find(providerName, targetName)
	if err != nil {
		return err
	}

	return s.provisioner.StopProject(project, target)
}
