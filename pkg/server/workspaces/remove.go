// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/telemetry"
	log "github.com/sirupsen/logrus"
)

func (s *WorkspaceService) RemoveWorkspace(ctx context.Context, workspaceId string) error {
	workspace, err := s.workspaceStore.Find(workspaceId)
	if err != nil {
		return ErrWorkspaceNotFound
	}

	log.Infof("Destroying workspace %s", workspace.Id)

	target, err := s.targetStore.Find(workspace.Target)
	if err != nil {
		return err
	}

	for _, project := range workspace.Projects {
		//	todo: go routines
		err := s.provisioner.DestroyProject(project, target)
		if err != nil {
			return err
		}
	}

	err = s.provisioner.DestroyWorkspace(workspace, target)
	if err != nil {
		return err
	}

	// Should not fail the whole operation if the API key cannot be revoked
	err = s.apiKeyService.Revoke(workspace.Id)
	if err != nil {
		log.Error(err)
	}

	for _, project := range workspace.Projects {
		err := s.apiKeyService.Revoke(fmt.Sprintf("%s/%s", workspace.Id, project.Name))
		if err != nil {
			// Should not fail the whole operation if the API key cannot be revoked
			log.Error(err)
		}
		projectLogger := s.loggerFactory.CreateProjectLogger(workspace.Id, project.Name, logs.LogSourceServer)
		err = projectLogger.Cleanup()
		if err != nil {
			// Should not fail the whole operation if the project logger cannot be cleaned up
			log.Error(err)
		}
	}

	logger := s.loggerFactory.CreateWorkspaceLogger(workspace.Id, logs.LogSourceServer)
	err = logger.Cleanup()
	if err != nil {
		// Should not fail the whole operation if the workspace logger cannot be cleaned up
		log.Error(err)
	}

	err = s.workspaceStore.Delete(workspace)

	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	cliId := telemetry.CliId(ctx)

	telemetryProps := telemetry.NewWorkspaceEventProps(ctx, workspace, target)
	event := telemetry.ServerEventWorkspaceDestroyed
	if err != nil {
		telemetryProps["error"] = err.Error()
		event = telemetry.ServerEventWorkspaceDestroyError
	}
	telemetryError := s.telemetryService.TrackServerEvent(event, cliId, telemetryProps)
	if telemetryError != nil {
		log.Trace(telemetryError)
	}

	return err
}

// ForceRemoveWorkspace ignores provider errors and makes sure the workspace is removed from storage.
func (s *WorkspaceService) ForceRemoveWorkspace(ctx context.Context, workspaceId string) error {
	workspace, err := s.workspaceStore.Find(workspaceId)
	if err != nil {
		return ErrWorkspaceNotFound
	}

	log.Infof("Destroying workspace %s", workspace.Id)

	target, _ := s.targetStore.Find(workspace.Target)

	for _, project := range workspace.Projects {
		//	todo: go routines
		err := s.provisioner.DestroyProject(project, target)
		if err != nil {
			log.Error(err)
		}
	}

	err = s.provisioner.DestroyWorkspace(workspace, target)
	if err != nil {
		log.Error(err)
	}

	err = s.apiKeyService.Revoke(workspace.Id)
	if err != nil {
		log.Error(err)
	}

	for _, project := range workspace.Projects {
		err := s.apiKeyService.Revoke(fmt.Sprintf("%s/%s", workspace.Id, project.Name))
		if err != nil {
			log.Error(err)
		}

		projectLogger := s.loggerFactory.CreateProjectLogger(workspace.Id, project.Name, logs.LogSourceServer)
		err = projectLogger.Cleanup()
		if err != nil {
			log.Error(err)
		}
	}

	err = s.workspaceStore.Delete(workspace)

	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	cliId := telemetry.CliId(ctx)

	telemetryProps := telemetry.NewWorkspaceEventProps(ctx, workspace, target)
	event := telemetry.ServerEventWorkspaceDestroyed
	if err != nil {
		telemetryProps["error"] = err.Error()
		event = telemetry.ServerEventWorkspaceDestroyError
	}
	telemetryError := s.telemetryService.TrackServerEvent(event, cliId, telemetryProps)
	if telemetryError != nil {
		log.Trace(telemetryError)
	}

	return err
}
