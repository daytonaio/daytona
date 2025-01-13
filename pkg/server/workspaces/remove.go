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

	w, err := s.workspaceStore.Find(ctx, workspaceId)
	if err != nil {
		return s.handleRemoveError(ctx, w, stores.ErrWorkspaceNotFound)
	}

	w.Name = util.AddDeletedToName(w.Name)

	err = s.workspaceStore.Save(ctx, w)
	if err != nil {
		return s.handleRemoveError(ctx, w, err)
	}

	err = s.revokeApiKey(ctx, workspaceId)
	if err != nil {
		return s.handleRemoveError(ctx, w, err)
	}

	metadata, err := s.workspaceMetadataStore.Find(ctx, workspaceId)
	if err != nil {
		return s.handleRemoveError(ctx, w, err)
	}

	err = s.workspaceMetadataStore.Delete(ctx, metadata)
	if err != nil {
		return s.handleRemoveError(ctx, w, err)
	}

	err = s.createJob(ctx, w.Id, w.Target.TargetConfig.ProviderInfo.RunnerId, models.JobActionDelete)
	if err != nil {
		return s.handleRemoveError(ctx, w, err)
	}

	err = s.workspaceStore.CommitTransaction(ctx)
	return s.handleRemoveError(ctx, w, err)
}

// ForceRemoveWorkspace ignores provider errors and makes sure the workspace is removed from storage.
func (s *WorkspaceService) ForceRemoveWorkspace(ctx context.Context, workspaceId string) error {
	var err error
	ctx, err = s.workspaceStore.BeginTransaction(ctx)
	if err != nil {
		return s.handleForceRemoveError(ctx, nil, err)
	}

	defer stores.RecoverAndRollback(ctx, s.workspaceStore)

	w, err := s.workspaceStore.Find(ctx, workspaceId)
	if err != nil {
		return s.handleForceRemoveError(ctx, w, stores.ErrWorkspaceNotFound)
	}

	w.Name = util.AddDeletedToName(w.Name)

	err = s.workspaceStore.Save(ctx, w)
	if err != nil {
		return s.handleForceRemoveError(ctx, w, err)
	}

	err = s.revokeApiKey(ctx, workspaceId)
	if err != nil {
		// Should not fail the whole operation if the API key cannot be revoked
		log.Error(err)
	}

	metadata, err := s.workspaceMetadataStore.Find(ctx, workspaceId)
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

	err = s.createJob(ctx, w.Id, w.Target.TargetConfig.ProviderInfo.RunnerId, models.JobActionForceDelete)
	if err != nil {
		return s.handleForceRemoveError(ctx, w, err)
	}

	err = s.workspaceStore.CommitTransaction(ctx)
	return s.handleForceRemoveError(ctx, w, err)
}

func (s *WorkspaceService) handleRemoveError(ctx context.Context, w *models.Workspace, err error) error {
	if err != nil {
		err = s.workspaceStore.RollbackTransaction(ctx, err)
	}

	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	clientId := telemetry.ClientId(ctx)

	eventName := telemetry.WorkspaceEventLifecycleDeleted
	if err != nil {
		eventName = telemetry.WorkspaceEventLifecycleDeletionFailed
	}
	event := telemetry.NewWorkspaceEvent(eventName, w, err, nil)

	telemetryError := s.trackTelemetryEvent(event, clientId)
	if telemetryError != nil {
		log.Trace(telemetryError)
	}

	return err
}

func (s *WorkspaceService) handleForceRemoveError(ctx context.Context, w *models.Workspace, err error) error {
	if err != nil {
		err = s.workspaceStore.RollbackTransaction(ctx, err)
	}

	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	clientId := telemetry.ClientId(ctx)

	eventName := telemetry.WorkspaceEventLifecycleForceDeleted
	if err != nil {
		eventName = telemetry.WorkspaceEventLifecycleForceDeletionFailed
	}
	event := telemetry.NewWorkspaceEvent(eventName, w, err, nil)

	telemetryError := s.trackTelemetryEvent(event, clientId)
	if telemetryError != nil {
		log.Trace(telemetryError)
	}

	return err
}
