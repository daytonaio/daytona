// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"context"

	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/telemetry"
	log "github.com/sirupsen/logrus"
)

func (s *TargetService) RemoveTarget(ctx context.Context, targetId string) error {
	target, err := s.targetStore.Find(targetId)
	if err != nil {
		return s.handleRemoveError(ctx, target, ErrTargetNotFound)
	}

	log.Infof("Destroying target %s", target.Id)

	targetConfig, err := s.targetConfigStore.Find(&provider.TargetConfigFilter{Name: &target.TargetConfig})
	if err != nil {
		return s.handleRemoveError(ctx, target, err)
	}

	err = s.provisioner.DestroyTarget(target, targetConfig)
	if err != nil {
		return s.handleRemoveError(ctx, target, err)
	}

	// Should not fail the whole operation if the API key cannot be revoked
	err = s.apiKeyService.Revoke(target.Id)
	if err != nil {
		log.Error(err)
	}

	logger := s.loggerFactory.CreateTargetLogger(target.Id, target.Name, logs.LogSourceServer)
	err = logger.Cleanup()
	if err != nil {
		// Should not fail the whole operation if the target logger cannot be cleaned up
		log.Error(err)
	}

	err = s.targetStore.Delete(target)

	return s.handleRemoveError(ctx, target, err)
}

// ForceRemoveTarget ignores provider errors and makes sure the target is removed from storage.
func (s *TargetService) ForceRemoveTarget(ctx context.Context, targetId string) error {
	target, err := s.targetStore.Find(targetId)
	if err != nil {
		return s.handleRemoveError(ctx, nil, ErrTargetNotFound)
	}

	log.Infof("Destroying target %s", target.Id)

	targetConfig, _ := s.targetConfigStore.Find(&provider.TargetConfigFilter{Name: &target.TargetConfig})

	err = s.provisioner.DestroyTarget(target, targetConfig)
	if err != nil {
		log.Error(err)
	}

	err = s.apiKeyService.Revoke(target.Id)
	if err != nil {
		log.Error(err)
	}

	err = s.targetStore.Delete(target)

	return s.handleRemoveError(ctx, target, err)
}

func (s *TargetService) handleRemoveError(ctx context.Context, target *target.Target, err error) error {
	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	clientId := telemetry.ClientId(ctx)

	telemetryProps := telemetry.NewTargetEventProps(ctx, target, nil)
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
