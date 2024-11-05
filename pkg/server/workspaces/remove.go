// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/daytonaio/daytona/pkg/workspace"
	log "github.com/sirupsen/logrus"
)

func (s *WorkspaceService) RemoveWorkspace(ctx context.Context, workspaceId string) error {
	ws, err := s.workspaceStore.Find(workspaceId)
	if err != nil {
		return s.handleRemoveError(ctx, &ws.Workspace, ErrWorkspaceNotFound)
	}

	log.Infof("Destroying workspace %s", ws.Name)

	target, err := s.targetStore.Find(&target.TargetFilter{IdOrName: &ws.TargetId})
	if err != nil {
		return s.handleRemoveError(ctx, &ws.Workspace, err)
	}

	err = s.provisioner.DestroyWorkspace(&ws.Workspace, &target.Target)
	if err != nil {
		return s.handleRemoveError(ctx, &ws.Workspace, err)
	}

	err = s.apiKeyService.Revoke(fmt.Sprintf("ws-%s", ws.Id))
	if err != nil {
		// Should not fail the whole operation if the API key cannot be revoked
		log.Error(err)
	}
	workspaceLogger := s.loggerFactory.CreateWorkspaceLogger(ws.Id, ws.Name, logs.LogSourceServer)
	err = workspaceLogger.Cleanup()
	if err != nil {
		// Should not fail the whole operation if the workspace logger cannot be cleaned up
		log.Error(err)
	}

	err = s.workspaceStore.Delete(&ws.Workspace)

	return s.handleRemoveError(ctx, &ws.Workspace, err)
}

// ForceRemoveWorkspace ignores provider errors and makes sure the workspace is removed from storage.
func (s *WorkspaceService) ForceRemoveWorkspace(ctx context.Context, workspaceId string) error {
	ws, err := s.workspaceStore.Find(workspaceId)
	if err != nil {
		return s.handleRemoveError(ctx, &ws.Workspace, ErrWorkspaceNotFound)
	}

	log.Infof("Destroying workspace %s", ws.Name)

	target, err := s.targetStore.Find(&target.TargetFilter{IdOrName: &ws.TargetId})
	if err != nil {
		return s.handleRemoveError(ctx, &ws.Workspace, err)
	}

	err = s.provisioner.DestroyWorkspace(&ws.Workspace, &target.Target)
	if err != nil {
		log.Error(err)
	}

	err = s.apiKeyService.Revoke(fmt.Sprintf("ws-%s", ws.Id))
	if err != nil {
		// Should not fail the whole operation if the API key cannot be revoked
		log.Error(err)
	}
	workspaceLogger := s.loggerFactory.CreateWorkspaceLogger(ws.Id, ws.Name, logs.LogSourceServer)
	err = workspaceLogger.Cleanup()
	if err != nil {
		// Should not fail the whole operation if the workspace logger cannot be cleaned up
		log.Error(err)
	}

	err = s.workspaceStore.Delete(&ws.Workspace)

	return s.handleRemoveError(ctx, &ws.Workspace, err)
}

func (s *WorkspaceService) handleRemoveError(ctx context.Context, w *workspace.Workspace, err error) error {
	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	clientId := telemetry.ClientId(ctx)

	telemetryProps := telemetry.NewWorkspaceEventProps(ctx, w)
	event := telemetry.ServerEventWorkspaceDestroyed
	if err != nil {
		telemetryProps["error"] = err.Error()
		event = telemetry.ServerEventWorkspaceDestroyError
	}
	telemetryError := s.telemetryService.TrackServerEvent(event, clientId, telemetryProps)
	if telemetryError != nil {
		log.Trace(err)
	}

	return err
}
