// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/daytonaio/daytona/pkg/telemetry"
	log "github.com/sirupsen/logrus"
)

func (s *WorkspaceService) StopWorkspace(ctx context.Context, workspaceId string) error {
	ws, err := s.workspaceStore.Find(workspaceId)
	if err != nil {
		return s.handleStopError(ctx, ws, stores.ErrWorkspaceNotFound)
	}

	err = s.createJob(ctx, ws.Id, models.JobActionStop)
	return s.handleStopError(ctx, ws, err)
}

func (s *WorkspaceService) handleStopError(ctx context.Context, w *models.Workspace, err error) error {
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
	telemetryError := s.trackTelemetryEvent(event, clientId, telemetryProps)
	if telemetryError != nil {
		log.Trace(err)
	}

	return err
}
