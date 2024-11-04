// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"context"
	"fmt"
	"io"

	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/daytonaio/daytona/pkg/workspace"
	log "github.com/sirupsen/logrus"
)

func (s *WorkspaceService) StartWorkspace(ctx context.Context, workspaceId string) error {
	ws, err := s.workspaceStore.Find(workspaceId)
	if err != nil {
		return s.handleStartError(ctx, &ws.Workspace, ErrWorkspaceNotFound)
	}

	target, err := s.targetStore.Find(&target.TargetFilter{IdOrName: &ws.TargetId})
	if err != nil {
		return s.handleStartError(ctx, &ws.Workspace, err)
	}

	workspaceLogger := s.loggerFactory.CreateWorkspaceLogger(ws.Id, ws.Name, logs.LogSourceServer)
	defer workspaceLogger.Close()

	workspaceToStart := ws.Workspace
	workspaceToStart.EnvVars = workspace.GetWorkspaceEnvVars(&ws.Workspace, workspace.WorkspaceEnvVarParams{
		ApiUrl:        s.serverApiUrl,
		ServerUrl:     s.serverUrl,
		ServerVersion: s.serverVersion,
		ClientId:      telemetry.ClientId(ctx),
	}, telemetry.TelemetryEnabled(ctx))

	err = s.startWorkspace(&workspaceToStart, target, workspaceLogger)
	if err != nil {
		return s.handleStartError(ctx, &ws.Workspace, err)
	}

	return s.handleStartError(ctx, &ws.Workspace, err)
}

func (s *WorkspaceService) startWorkspace(w *workspace.Workspace, target *target.Target, logger io.Writer) error {
	logger.Write([]byte(fmt.Sprintf("Starting workspace %s\n", w.Name)))

	err := s.provisioner.StartWorkspace(w, target)
	if err != nil {
		return err
	}

	logger.Write([]byte(fmt.Sprintf("Workspace %s started\n", w.Name)))

	return nil
}

func (s *WorkspaceService) handleStartError(ctx context.Context, w *workspace.Workspace, err error) error {
	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	clientId := telemetry.ClientId(ctx)

	telemetryProps := telemetry.NewWorkspaceEventProps(ctx, w)
	event := telemetry.ServerEventWorkspaceStarted
	if err != nil {
		telemetryProps["error"] = err.Error()
		event = telemetry.ServerEventWorkspaceStartError
	}
	telemetryError := s.telemetryService.TrackServerEvent(event, clientId, telemetryProps)
	if telemetryError != nil {
		log.Trace(err)
	}

	return err
}
