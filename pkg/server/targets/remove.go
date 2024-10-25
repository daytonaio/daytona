// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/telemetry"
	log "github.com/sirupsen/logrus"
)

func (s *TargetService) RemoveTarget(ctx context.Context, targetId string) error {
	target, err := s.targetStore.Find(targetId)
	if err != nil {
		return ErrTargetNotFound
	}

	log.Infof("Destroying target %s", target.Id)

	targetConfig, err := s.targetConfigStore.Find(&provider.TargetConfigFilter{Name: &target.TargetConfig})
	if err != nil {
		return err
	}

	for _, project := range target.Projects {
		//	todo: go routines
		err := s.provisioner.DestroyProject(project, targetConfig)
		if err != nil {
			return err
		}
	}

	err = s.provisioner.DestroyTarget(target, targetConfig)
	if err != nil {
		return err
	}

	// Should not fail the whole operation if the API key cannot be revoked
	err = s.apiKeyService.Revoke(target.Id)
	if err != nil {
		log.Error(err)
	}

	for _, project := range target.Projects {
		err := s.apiKeyService.Revoke(fmt.Sprintf("%s/%s", target.Id, project.Name))
		if err != nil {
			// Should not fail the whole operation if the API key cannot be revoked
			log.Error(err)
		}
		projectLogger := s.loggerFactory.CreateProjectLogger(target.Id, project.Name, logs.LogSourceServer)
		err = projectLogger.Cleanup()
		if err != nil {
			// Should not fail the whole operation if the project logger cannot be cleaned up
			log.Error(err)
		}
	}

	logger := s.loggerFactory.CreateTargetLogger(target.Id, logs.LogSourceServer)
	err = logger.Cleanup()
	if err != nil {
		// Should not fail the whole operation if the target logger cannot be cleaned up
		log.Error(err)
	}

	err = s.targetStore.Delete(target)

	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	clientId := telemetry.ClientId(ctx)

	telemetryProps := telemetry.NewTargetEventProps(ctx, target, targetConfig)
	event := telemetry.ServerEventTargetDestroyed
	if err != nil {
		telemetryProps["error"] = err.Error()
		event = telemetry.ServerEventTargetDestroyError
	}
	telemetryError := s.telemetryService.TrackServerEvent(event, clientId, telemetryProps)
	if telemetryError != nil {
		log.Trace(telemetryError)
	}

	return err
}

// ForceRemoveTarget ignores provider errors and makes sure the target is removed from storage.
func (s *TargetService) ForceRemoveTarget(ctx context.Context, targetId string) error {
	target, err := s.targetStore.Find(targetId)
	if err != nil {
		return ErrTargetNotFound
	}

	log.Infof("Destroying target %s", target.Id)

	targetConfig, _ := s.targetConfigStore.Find(&provider.TargetConfigFilter{Name: &target.TargetConfig})

	for _, project := range target.Projects {
		//	todo: go routines
		err := s.provisioner.DestroyProject(project, targetConfig)
		if err != nil {
			log.Error(err)
		}
	}

	err = s.provisioner.DestroyTarget(target, targetConfig)
	if err != nil {
		log.Error(err)
	}

	err = s.apiKeyService.Revoke(target.Id)
	if err != nil {
		log.Error(err)
	}

	for _, project := range target.Projects {
		err := s.apiKeyService.Revoke(fmt.Sprintf("%s/%s", target.Id, project.Name))
		if err != nil {
			log.Error(err)
		}

		projectLogger := s.loggerFactory.CreateProjectLogger(target.Id, project.Name, logs.LogSourceServer)
		err = projectLogger.Cleanup()
		if err != nil {
			log.Error(err)
		}
	}

	err = s.targetStore.Delete(target)

	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	clientId := telemetry.ClientId(ctx)

	telemetryProps := telemetry.NewTargetEventProps(ctx, target, targetConfig)
	event := telemetry.ServerEventTargetDestroyed
	if err != nil {
		telemetryProps["error"] = err.Error()
		event = telemetry.ServerEventTargetDestroyError
	}
	telemetryError := s.telemetryService.TrackServerEvent(event, clientId, telemetryProps)
	if telemetryError != nil {
		log.Trace(telemetryError)
	}

	return err
}
