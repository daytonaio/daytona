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

func (s *WorkspaceService) Delete(ctx context.Context, workspaceId string) error {
	var err error
	ctx, err = s.workspaceStore.BeginTransaction(ctx)
	if err != nil {
		return s.handleDeleteError(ctx, nil, err)
	}

	defer stores.RecoverAndRollback(ctx, s.workspaceStore)

	w, err := s.workspaceStore.Find(ctx, workspaceId)
	if err != nil {
		return s.handleDeleteError(ctx, w, stores.ErrWorkspaceNotFound)
	}

	w.Name = util.AddDeletedToName(w.Name)

	err = s.workspaceStore.Save(ctx, w)
	if err != nil {
		return s.handleDeleteError(ctx, w, err)
	}

	err = s.deleteApiKey(ctx, workspaceId)
	if err != nil {
		return s.handleDeleteError(ctx, w, err)
	}

	metadata, err := s.workspaceMetadataStore.Find(ctx, workspaceId)
	if err == nil {
		err = s.workspaceMetadataStore.Delete(ctx, metadata)
		if err != nil {
			return s.handleDeleteError(ctx, w, err)
		}
	}

	err = s.createJob(ctx, w.Id, w.Target.TargetConfig.ProviderInfo.RunnerId, models.JobActionDelete)
	if err != nil {
		return s.handleDeleteError(ctx, w, err)
	}

	err = s.workspaceStore.CommitTransaction(ctx)
	return s.handleDeleteError(ctx, w, err)
}

// ForceDelete ignores provider errors and makes sure the workspace is removed from storage.
func (s *WorkspaceService) ForceDelete(ctx context.Context, workspaceId string) error {
	var err error
	ctx, err = s.workspaceStore.BeginTransaction(ctx)
	if err != nil {
		return s.handleForceDeleteError(ctx, nil, err)
	}

	defer stores.RecoverAndRollback(ctx, s.workspaceStore)

	w, err := s.workspaceStore.Find(ctx, workspaceId)
	if err != nil {
		return s.handleForceDeleteError(ctx, w, stores.ErrWorkspaceNotFound)
	}

	w.Name = util.AddDeletedToName(w.Name)

	err = s.workspaceStore.Save(ctx, w)
	if err != nil {
		return s.handleForceDeleteError(ctx, w, err)
	}

	err = s.deleteApiKey(ctx, workspaceId)
	if err != nil {
		log.Error(err)
	}

	metadata, err := s.workspaceMetadataStore.Find(ctx, workspaceId)
	if err == nil {
		err = s.workspaceMetadataStore.Delete(ctx, metadata)
		if err != nil {
			log.Error(err)
		}
	}

	err = s.createJob(ctx, w.Id, w.Target.TargetConfig.ProviderInfo.RunnerId, models.JobActionForceDelete)
	if err != nil {
		return s.handleForceDeleteError(ctx, w, err)
	}

	err = s.workspaceStore.CommitTransaction(ctx)
	return s.handleForceDeleteError(ctx, w, err)
}

func (s *WorkspaceService) handleDeleteError(ctx context.Context, w *models.Workspace, err error) error {
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

func (s *WorkspaceService) handleForceDeleteError(ctx context.Context, w *models.Workspace, err error) error {
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
