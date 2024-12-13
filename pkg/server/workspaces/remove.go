// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"context"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/daytonaio/daytona/pkg/telemetry"
	log "github.com/sirupsen/logrus"
)

func (s *WorkspaceService) RemoveWorkspace(ctx context.Context, workspaceId string) error {
	var err error
	ctx, err = s.workspaceStore.BeginTransaction(ctx)
	if err != nil {
		return s.handleRemoveError(ctx, nil, err)
	}

	defer stores.RecoverAndRollback(ctx, s.workspaceStore)

	ws, err := s.workspaceStore.Find(ctx, workspaceId)
	if err != nil {
		return s.handleRemoveError(ctx, ws, stores.ErrWorkspaceNotFound)
	}

	ws.Name = util.AddDeletedToName(ws.Name)

	err = s.workspaceStore.Save(ctx, ws)
	if err != nil {
		return s.handleRemoveError(ctx, ws, err)
	}

	err = s.revokeApiKey(ctx, workspaceId)
	if err != nil {
		return s.handleRemoveError(ctx, ws, err)
	}

	metadata, err := s.workspaceMetadataStore.Find(ctx, &stores.WorkspaceMetadataFilter{WorkspaceId: &workspaceId})
	if err != nil {
		return s.handleRemoveError(ctx, ws, err)
	}

	err = s.workspaceMetadataStore.Delete(ctx, metadata)
	if err != nil {
		return s.handleRemoveError(ctx, ws, err)
	}

	err = s.createJob(ctx, ws.Id, models.JobActionDelete)
	if err != nil {
		return s.handleRemoveError(ctx, ws, err)
	}

	err = s.workspaceStore.CommitTransaction(ctx)
	return s.handleRemoveError(ctx, ws, err)
}

// ForceRemoveWorkspace ignores provider errors and makes sure the workspace is removed from storage.
func (s *WorkspaceService) ForceRemoveWorkspace(ctx context.Context, workspaceId string) error {
	var err error
	ctx, err = s.workspaceStore.BeginTransaction(ctx)
	if err != nil {
		return s.handleRemoveError(ctx, nil, err)
	}

	defer stores.RecoverAndRollback(ctx, s.workspaceStore)

	ws, err := s.workspaceStore.Find(ctx, workspaceId)
	if err != nil {
		return s.handleRemoveError(ctx, ws, stores.ErrWorkspaceNotFound)
	}

	ws.Name = util.AddDeletedToName(ws.Name)

	err = s.workspaceStore.Save(ctx, ws)
	if err != nil {
		return s.handleRemoveError(ctx, ws, err)
	}

	err = s.revokeApiKey(ctx, workspaceId)
	if err != nil {
		// Should not fail the whole operation if the API key cannot be revoked
		log.Error(err)
	}

	metadata, err := s.workspaceMetadataStore.Find(ctx, &stores.WorkspaceMetadataFilter{WorkspaceId: &workspaceId})
	if err != nil {
		// Should not fail the whole operation if the metadata cannot be found
		log.Error(err)
	} else {
		err = s.workspaceMetadataStore.Delete(ctx, metadata)
		if err != nil {
			// Should not fail the whole operation if the metadata cannot be deleted
			log.Error(err)
		}
	}

	err = s.createJob(ctx, ws.Id, models.JobActionForceDelete)
	if err != nil {
		return s.handleRemoveError(ctx, ws, err)
	}

	err = s.workspaceStore.CommitTransaction(ctx)
	return s.handleRemoveError(ctx, ws, err)
}

func (s *WorkspaceService) handleRemoveError(ctx context.Context, w *models.Workspace, err error) error {
	if err != nil {
		err = s.workspaceStore.RollbackTransaction(ctx, err)
	}

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
