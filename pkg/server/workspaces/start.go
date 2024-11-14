// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"context"
	"fmt"
	"io"

	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/daytonaio/daytona/pkg/views"
	log "github.com/sirupsen/logrus"
)

func (s *WorkspaceService) StartWorkspace(ctx context.Context, workspaceId string) error {
	ws, err := s.workspaceStore.Find(workspaceId)
	if err != nil {
		return s.handleStartError(ctx, ws, ErrWorkspaceNotFound)
	}

	workspaceLogger := s.loggerFactory.CreateWorkspaceLogger(ws.Id, ws.Name, logs.LogSourceServer)
	defer workspaceLogger.Close()

	workspaceToStart := ws
	workspaceToStart.EnvVars = GetWorkspaceEnvVars(ws, WorkspaceEnvVarParams{
		ApiUrl:        s.serverApiUrl,
		ServerUrl:     s.serverUrl,
		ServerVersion: s.serverVersion,
		ClientId:      telemetry.ClientId(ctx),
	}, telemetry.TelemetryEnabled(ctx))

	err = s.startWorkspace(workspaceToStart, workspaceLogger)
	if err != nil {
		return s.handleStartError(ctx, ws, err)
	}

	return s.handleStartError(ctx, ws, err)
}

func (s *WorkspaceService) startWorkspace(w *models.Workspace, logger io.Writer) error {
	logger.Write([]byte(fmt.Sprintf("Starting workspace %s\n", w.Name)))

	err := s.provisioner.StartWorkspace(w)
	if err != nil {
		return err
	}

	logger.Write([]byte(views.GetPrettyLogLine(fmt.Sprintf("Workspace %s started", w.Name))))

	return nil
}

func (s *WorkspaceService) handleStartError(ctx context.Context, w *models.Workspace, err error) error {
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
	telemetryError := s.trackTelemetryEvent(event, clientId, telemetryProps)
	if telemetryError != nil {
		log.Trace(err)
	}

	return err
}
