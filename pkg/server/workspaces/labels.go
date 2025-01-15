// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/telemetry"
	log "github.com/sirupsen/logrus"
)

func (s *WorkspaceService) UpdateLabels(ctx context.Context, workspaceId string, labels map[string]string) (*services.WorkspaceDTO, error) {
	w, err := s.workspaceStore.Find(ctx, workspaceId)
	if err != nil {
		return nil, s.handleUpdateLabelsError(ctx, w, err)
	}

	w.Labels = labels

	if err := s.workspaceStore.Save(ctx, w); err != nil {
		return nil, s.handleUpdateLabelsError(ctx, w, err)
	}

	return &services.WorkspaceDTO{Workspace: *w, State: w.GetState()}, s.handleUpdateLabelsError(ctx, w, nil)
}

func (s *WorkspaceService) handleUpdateLabelsError(ctx context.Context, w *models.Workspace, err error) error {
	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	clientId := telemetry.ClientId(ctx)

	eventName := telemetry.WorkspaceEventLabelsUpdated
	if err != nil {
		eventName = telemetry.WorkspaceEventLabelsUpdateFailed
	}
	event := telemetry.NewWorkspaceEvent(eventName, w, err, nil)

	telemetryError := s.trackTelemetryEvent(event, clientId)
	if telemetryError != nil {
		log.Trace(telemetryError)
	}

	return err
}
