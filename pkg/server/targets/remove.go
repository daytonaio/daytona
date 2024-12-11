// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"context"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/daytonaio/daytona/pkg/telemetry"
	log "github.com/sirupsen/logrus"
)

func (s *TargetService) RemoveTarget(ctx context.Context, targetId string) error {
	var err error
	ctx, err = s.targetStore.BeginTransaction(ctx)
	if err != nil {
		return s.handleRemoveError(ctx, nil, err)
	}

	defer stores.RecoverAndRollback(ctx, s.targetStore)

	t, err := s.targetStore.Find(ctx, &stores.TargetFilter{IdOrName: &targetId})
	if err != nil {
		return s.handleRemoveError(ctx, t, stores.ErrTargetNotFound)
	}

	t.Name = util.AddDeletedToName(t.Name)

	err = s.targetStore.Save(ctx, t)
	if err != nil {
		return s.handleRemoveError(ctx, t, err)
	}

	err = s.revokeApiKey(ctx, targetId)
	if err != nil {
		return s.handleRemoveError(ctx, t, err)
	}

	metadata, err := s.targetMetadataStore.Find(ctx, &stores.TargetMetadataFilter{TargetId: &targetId})
	if err != nil {
		return s.handleRemoveError(ctx, t, err)
	}

	err = s.targetMetadataStore.Delete(ctx, metadata)
	if err != nil {
		return s.handleRemoveError(ctx, t, err)
	}

	err = s.createJob(ctx, t.Id, models.JobActionDelete)
	if err != nil {
		return s.handleRemoveError(ctx, t, err)
	}

	err = s.targetStore.CommitTransaction(ctx)
	return s.handleRemoveError(ctx, t, err)
}

// ForceRemoveTarget ignores provider errors and makes sure the target is removed from storage.
func (s *TargetService) ForceRemoveTarget(ctx context.Context, targetId string) error {
	var err error
	ctx, err = s.targetStore.BeginTransaction(ctx)
	if err != nil {
		return s.handleRemoveError(ctx, nil, err)
	}

	defer stores.RecoverAndRollback(ctx, s.targetStore)

	t, err := s.targetStore.Find(ctx, &stores.TargetFilter{IdOrName: &targetId})
	if err != nil {
		return s.handleRemoveError(ctx, nil, stores.ErrTargetNotFound)
	}

	t.Name = util.AddDeletedToName(t.Name)

	err = s.targetStore.Save(ctx, t)
	if err != nil {
		return s.handleRemoveError(ctx, t, err)
	}

	err = s.revokeApiKey(ctx, targetId)
	if err != nil {
		// Should not fail the whole operation if the API key cannot be revoked
		log.Error(err)
	}

	metadata, err := s.targetMetadataStore.Find(ctx, &stores.TargetMetadataFilter{TargetId: &targetId})
	if err != nil {
		// Should not fail the whole operation if the metadata cannot be found
		log.Error(err)
	} else {
		err = s.targetMetadataStore.Delete(ctx, metadata)
		if err != nil {
			// Should not fail the whole operation if the metadata cannot be deleted
			log.Error(err)
		}
	}

	err = s.createJob(ctx, t.Id, models.JobActionForceDelete)
	if err != nil {
		return s.handleRemoveError(ctx, t, err)
	}

	err = s.targetStore.CommitTransaction(ctx)
	return s.handleRemoveError(ctx, t, err)
}

func (s *TargetService) handleRemoveError(ctx context.Context, target *models.Target, err error) error {
	if err != nil {
		err = s.targetStore.RollbackTransaction(ctx, err)
	}

	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	clientId := telemetry.ClientId(ctx)

	telemetryProps := telemetry.NewTargetEventProps(ctx, target)
	event := telemetry.ServerEventTargetDestroyed
	if err != nil {
		telemetryProps["error"] = err.Error()
		event = telemetry.ServerEventTargetDestroyError
	}
	telemetryError := s.telemetryService.TrackServerEvent(event, clientId, telemetryProps)
	if telemetryError != nil {
		log.Trace(err)
	}

	return err
}
