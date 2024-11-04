// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"context"

	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/telemetry"
	log "github.com/sirupsen/logrus"
)

func (s *TargetService) StopTarget(ctx context.Context, targetId string) error {
	target, err := s.targetStore.Find(targetId)
	if err != nil {
		return s.handleStopError(ctx, nil, ErrTargetNotFound)
	}

	targetConfig, err := s.targetConfigStore.Find(&provider.TargetConfigFilter{Name: &target.TargetConfig})
	if err != nil {
		return s.handleStopError(ctx, target, err)
	}

	err = s.provisioner.StopTarget(target, targetConfig)

	return s.handleStopError(ctx, target, err)
}

func (s *TargetService) handleStopError(ctx context.Context, target *target.Target, err error) error {
	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	clientId := telemetry.ClientId(ctx)

	telemetryProps := telemetry.NewTargetEventProps(ctx, target, nil)
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
