// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/daytonaio/daytona/pkg/telemetry"
	log "github.com/sirupsen/logrus"
)

func (s *TargetService) RestartTarget(ctx context.Context, targetId string) error {
	target, err := s.targetStore.Find(ctx, &stores.TargetFilter{IdOrName: &targetId})
	if err != nil {
		return s.handleRestartError(ctx, nil, stores.ErrTargetNotFound)
	}

	if target.TargetConfig.ProviderInfo.AgentlessTarget {
		return s.handleRestartError(ctx, target, services.ErrAgentlessTarget)
	}

	err = s.createJob(ctx, target.Id, target.TargetConfig.ProviderInfo.RunnerId, models.JobActionRestart)
	return s.handleRestartError(ctx, target, err)
}

func (s *TargetService) handleRestartError(ctx context.Context, target *models.Target, err error) error {
	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	clientId := telemetry.ClientId(ctx)

	telemetryProps := telemetry.NewTargetEventProps(ctx, target)
	event := telemetry.ServerEventTargetRestarted
	if err != nil {
		telemetryProps["error"] = err.Error()
		event = telemetry.ServerEventTargetRestartError
	}
	telemetryError := s.telemetryService.TrackServerEvent(event, clientId, telemetryProps)
	if telemetryError != nil {
		log.Trace(telemetryError)
	}

	return err
}
