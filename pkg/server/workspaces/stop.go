// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"context"
	"time"

	"github.com/daytonaio/daytona/pkg/telemetry"
	log "github.com/sirupsen/logrus"
)

func (s *WorkspaceService) StopWorkspace(ctx context.Context, workspaceId string) error {
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

	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	cliId := telemetry.CliId(ctx)

	telemetryProps := telemetry.NewWorkspaceEventProps(ctx, workspace, target)
	event := telemetry.ServerEventWorkspaceStopped
	if err != nil {
		telemetryProps["error"] = err.Error()
		event = telemetry.ServerEventWorkspaceStopError
	}
	telemetryError := s.telemetryService.TrackServerEvent(event, cliId, telemetryProps)
	if telemetryError != nil {
		log.Trace(telemetryError)
	}

	if err != nil {
		return err
	}

	return s.workspaceStore.Save(workspace)
}

func (s *WorkspaceService) StopProject(ctx context.Context, workspaceId, projectName string) error {
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
