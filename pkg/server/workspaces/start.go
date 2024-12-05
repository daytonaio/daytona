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

func (s *WorkspaceService) StartWorkspace(ctx context.Context, workspaceId string) error {
	ws, err := s.workspaceStore.Find(ctx, workspaceId)
	if err != nil {
		return s.handleStartError(ctx, ws, stores.ErrWorkspaceNotFound)
	}

	err = s.createJob(ctx, ws.Id, models.JobActionStart)
	if err != nil {
		return s.handleStartError(ctx, ws, err)
	}

	return s.handleStartError(ctx, ws, err)
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
