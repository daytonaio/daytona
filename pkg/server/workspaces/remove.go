// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/daytonaio/daytona/pkg/telemetry"
	log "github.com/sirupsen/logrus"
)

func (s *WorkspaceService) RemoveWorkspace(ctx context.Context, workspaceId string) error {
	ws, err := s.workspaceStore.Find(workspaceId)
	if err != nil {
		return s.handleRemoveError(ctx, ws, stores.ErrWorkspaceNotFound)
	}

	ws.Name = util.AddDeletedToName(ws.Name)

	err = s.workspaceStore.Save(ws)
	if err != nil {
		return s.handleRemoveError(ctx, ws, err)
	}

	err = s.createJob(ctx, ws.Id, models.JobActionDelete)
	return s.handleRemoveError(ctx, ws, err)
}

// ForceRemoveWorkspace ignores provider errors and makes sure the workspace is removed from storage.
func (s *WorkspaceService) ForceRemoveWorkspace(ctx context.Context, workspaceId string) error {
	ws, err := s.workspaceStore.Find(workspaceId)
	if err != nil {
		return s.handleRemoveError(ctx, ws, stores.ErrWorkspaceNotFound)
	}

	ws.Name = util.AddDeletedToName(ws.Name)

	err = s.workspaceStore.Save(ws)
	if err != nil {
		return s.handleRemoveError(ctx, ws, err)
	}

	err = s.createJob(ctx, ws.Id, models.JobActionForceDelete)
	return s.handleRemoveError(ctx, ws, err)
}

func (s *WorkspaceService) HandleSuccessfulRemoval(ctx context.Context, workspaceId string) error {
	err := s.revokeApiKey(ctx, fmt.Sprintf("ws-%s", workspaceId))
	if err != nil {
		// Should not fail the whole operation if the API key cannot be revoked
		log.Error(err)
	}

	ws, err := s.workspaceStore.Find(workspaceId)
	if err != nil {
		return s.handleRemoveError(ctx, ws, stores.ErrWorkspaceNotFound)
	}

	metadata, err := s.workspaceMetadataStore.Find(&stores.WorkspaceMetadataFilter{WorkspaceId: &workspaceId})
	if err != nil {
		return s.handleRemoveError(ctx, ws, err)
	}

	err = s.workspaceMetadataStore.Delete(metadata)
	if err != nil {
		return s.handleRemoveError(ctx, ws, err)
	}

	err = s.workspaceStore.Delete(ws)
	return s.handleRemoveError(ctx, ws, err)
}

func (s *WorkspaceService) handleRemoveError(ctx context.Context, w *models.Workspace, err error) error {
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
	telemetryError := s.trackTelemetryEvent(event, clientId, telemetryProps)
	if telemetryError != nil {
		log.Trace(err)
	}

	return err
}
