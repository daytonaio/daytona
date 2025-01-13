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
	w, err := s.workspaceStore.Find(ctx, workspaceId)
	if err != nil {
		return s.handleStopError(ctx, w, stores.ErrWorkspaceNotFound)
	}

	err = s.createJob(ctx, w.Id, w.Target.TargetConfig.ProviderInfo.RunnerId, models.JobActionStop)
	if err != nil {
		return s.handleStopError(ctx, w, err)
	}

	return s.handleStopError(ctx, w, err)
}

func (s *WorkspaceService) handleStopError(ctx context.Context, w *models.Workspace, err error) error {
	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	clientId := telemetry.ClientId(ctx)

	eventName := telemetry.WorkspaceEventLifecycleStopped
	if err != nil {
		eventName = telemetry.WorkspaceEventLifecycleStopFailed
	}
	event := telemetry.NewWorkspaceEvent(eventName, w, err, nil)

	telemetryError := s.trackTelemetryEvent(event, clientId)
	if telemetryError != nil {
		log.Trace(telemetryError)
	}

	return err
}
