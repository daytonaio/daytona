// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/telemetry"
	log "github.com/sirupsen/logrus"
)

func (s *TargetService) StopTarget(ctx context.Context, targetId string) error {
	target, err := s.targetStore.Find(&TargetFilter{IdOrName: &targetId})
	if err != nil {
		return s.handleStopError(ctx, nil, ErrTargetNotFound)
	}

	err = s.provisioner.StopTarget(target)

	return s.handleStopError(ctx, target, err)
}

func (s *TargetService) handleStopError(ctx context.Context, target *models.Target, err error) error {
	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	clientId := telemetry.ClientId(ctx)

	telemetryProps := telemetry.NewTargetEventProps(ctx, target)
	event := telemetry.ServerEventTargetStopped
	if err != nil {
		telemetryProps["error"] = err.Error()
		event = telemetry.ServerEventTargetStopError
	}
	telemetryError := s.telemetryService.TrackServerEvent(event, clientId, telemetryProps)
	if telemetryError != nil {
		log.Trace(telemetryError)
	}

	return err
}
