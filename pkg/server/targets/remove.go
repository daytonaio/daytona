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
	t, err := s.targetStore.Find(&stores.TargetFilter{IdOrName: &targetId})
	if err != nil {
		return s.handleRemoveError(ctx, t, stores.ErrTargetNotFound)
	}

	t.Name = util.AddDeletedToName(t.Name)

	err = s.targetStore.Save(t)
	if err != nil {
		return s.handleRemoveError(ctx, t, err)
	}

	err = s.createJob(ctx, t.Id, models.JobActionDelete)
	return s.handleRemoveError(ctx, t, err)
}

// ForceRemoveTarget ignores provider errors and makes sure the target is removed from storage.
func (s *TargetService) ForceRemoveTarget(ctx context.Context, targetId string) error {
	t, err := s.targetStore.Find(&stores.TargetFilter{IdOrName: &targetId})
	if err != nil {
		return s.handleRemoveError(ctx, nil, stores.ErrTargetNotFound)
	}

	t.Name = util.AddDeletedToName(t.Name)

	err = s.targetStore.Save(t)
	if err != nil {
		return s.handleRemoveError(ctx, t, err)
	}

	err = s.createJob(ctx, t.Id, models.JobActionForceDelete)
	return s.handleRemoveError(ctx, t, err)
}

func (s *TargetService) HandleSuccessfulRemoval(ctx context.Context, targetId string) error {
	err := s.revokeApiKey(ctx, targetId)
	if err != nil {
		// Should not fail the whole operation if the API key cannot be revoked
		log.Error(err)
	}

	t, err := s.targetStore.Find(&stores.TargetFilter{IdOrName: &targetId})
	if err != nil {
		return s.handleRemoveError(ctx, t, stores.ErrTargetNotFound)
	}

	metadata, err := s.targetMetadataStore.Find(&stores.TargetMetadataFilter{TargetId: &targetId})
	if err != nil {
		return s.handleRemoveError(ctx, t, err)
	}

	err = s.targetMetadataStore.Delete(metadata)
	if err != nil {
		return s.handleRemoveError(ctx, t, err)
	}

	err = s.targetStore.Delete(t)
	return s.handleRemoveError(ctx, t, err)
}

func (s *TargetService) handleRemoveError(ctx context.Context, target *models.Target, err error) error {
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
