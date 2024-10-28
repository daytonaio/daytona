// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"context"
	"time"

	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/daytonaio/daytona/pkg/workspace"
	log "github.com/sirupsen/logrus"
)

func (s *WorkspaceService) StopWorkspace(ctx context.Context, workspaceId string) error {
	ws, err := s.workspaceStore.Find(workspaceId)
	if err != nil {
		return s.handleStopError(ctx, ws, ErrWorkspaceNotFound)
	}

	target, err := s.targetStore.Find(ws.TargetId)
	if err != nil {
		return s.handleStopError(ctx, ws, err)
	}

	targetConfig, err := s.targetConfigStore.Find(&provider.TargetConfigFilter{Name: &target.TargetConfig})
	if err != nil {
		return s.handleStopError(ctx, ws, err)
	}

	//	todo: go routines
	err = s.provisioner.StopWorkspace(ws, targetConfig)
	if err != nil {
		return s.handleStopError(ctx, ws, err)
	}
	if ws.State != nil {
		ws.State.Uptime = 0
		ws.State.UpdatedAt = time.Now().Format(time.RFC1123)
	}

	err = s.workspaceStore.Save(ws)

	return s.handleStopError(ctx, ws, err)
}

func (s *WorkspaceService) handleStopError(ctx context.Context, w *workspace.Workspace, err error) error {
	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	clientId := telemetry.ClientId(ctx)

	telemetryProps := telemetry.NewWorkspaceEventProps(ctx, w)
	event := telemetry.ServerEventWorkspaceStopped
	if err != nil {
		telemetryProps["error"] = err.Error()
		event = telemetry.ServerEventWorkspaceStopError
	}
	telemetryError := s.telemetryService.TrackServerEvent(event, clientId, telemetryProps)
	if telemetryError != nil {
		log.Trace(err)
	}

	return err
}
