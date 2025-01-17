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

func (s *WorkspaceService) Start(ctx context.Context, workspaceId string) error {
	w, err := s.workspaceStore.Find(ctx, workspaceId)
	if err != nil {
		return s.handleStartError(ctx, w, stores.ErrWorkspaceNotFound)
	}

	err = s.createJob(ctx, w.Id, w.Target.TargetConfig.ProviderInfo.RunnerId, models.JobActionStart)
	if err != nil {
		return s.handleStartError(ctx, w, err)
	}

	return s.handleStartError(ctx, w, err)
}

func (s *WorkspaceService) handleStartError(ctx context.Context, w *models.Workspace, err error) error {
	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	clientId := telemetry.ClientId(ctx)

	eventName := telemetry.WorkspaceEventLifecycleStarted
	if err != nil {
		eventName = telemetry.WorkspaceEventLifecycleStartFailed
	}
	event := telemetry.NewWorkspaceEvent(eventName, w, err, nil)

	telemetryError := s.trackTelemetryEvent(event, clientId)
	if telemetryError != nil {
		log.Trace(telemetryError)
	}

	return err
}
