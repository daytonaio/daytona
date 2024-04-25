// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"time"
)

func (s *WorkspaceService) StopWorkspace(workspaceId string) error {
	workspace, err := s.WorkspaceStore.Find(workspaceId)
	if err != nil {
		return ErrWorkspaceNotFound
	}

	target, err := s.TargetStore.Find(workspace.Target)
	if err != nil {
		return err
	}

	for _, project := range workspace.Projects {
		//	todo: go routines
		err := s.Provisioner.StopProject(project, target)
		if err != nil {
			return err
		}
		if project.State != nil {
			project.State.Uptime = 0
			project.State.UpdatedAt = time.Now().Format(time.RFC1123)
		}
	}

	err = s.Provisioner.StopWorkspace(workspace, target)
	if err != nil {
		return err
	}

	return s.WorkspaceStore.Save(workspace)
}

func (s *WorkspaceService) StopProject(workspaceId, projectName string) error {
	w, err := s.WorkspaceStore.Find(workspaceId)
	if err != nil {
		return ErrWorkspaceNotFound
	}

	project, err := w.GetProject(projectName)
	if err != nil {
		return ErrProjectNotFound
	}

	target, err := s.TargetStore.Find(w.Target)
	if err != nil {
		return err
	}

	err = s.Provisioner.StopProject(project, target)
	if err != nil {
		return err
	}

	if project.State != nil {
		project.State.Uptime = 0
		project.State.UpdatedAt = time.Now().Format(time.RFC1123)
	}

	return s.WorkspaceStore.Save(w)
}
