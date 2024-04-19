// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"time"

	"github.com/daytonaio/daytona/pkg/telemetry"
)

func (s *WorkspaceService) StopWorkspace(workspaceId string) error {
	workspace, err := s.workspaceStore.Find(workspaceId)
	if err != nil {
		return ErrWorkspaceNotFound
	}

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
		if project.State != nil {
			project.State.Uptime = 0
			project.State.UpdatedAt = time.Now().Format(time.RFC1123)
		}
	}

	err = s.provisioner.StopWorkspace(workspace, target)

	telemetryProps := telemetry.NewWorkspaceEventProps(workspace, target)
	event := telemetry.ServerEventWorkspaceStopped
	if err != nil {
		telemetryProps["error"] = err.Error()
		event = telemetry.ServerEventWorkspaceStopError
	}
	s.telemetryService.TrackServerEvent(event, workspaceId, telemetryProps)

	if err != nil {
		return err
	}

	return s.workspaceStore.Save(workspace)
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

	err = s.provisioner.StopProject(project, target)
	if err != nil {
		return err
	}

	if project.State != nil {
		project.State.Uptime = 0
		project.State.UpdatedAt = time.Now().Format(time.RFC1123)
	}

	return s.workspaceStore.Save(w)
}
