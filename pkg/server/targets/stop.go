// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"context"

	db_dto "github.com/daytonaio/daytona/pkg/db/dto"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/telemetry"
	log "github.com/sirupsen/logrus"
)

func (s *TargetService) StopTarget(ctx context.Context, targetId string) error {
	target, err := s.targetStore.Find(&target.TargetFilter{IdOrName: &targetId})
	if err != nil {
		return s.handleStopError(ctx, nil, ErrTargetNotFound)
	}

	err = s.provisioner.StopTarget(db_dto.ViewToTarget(target))

	return s.handleStopError(ctx, db_dto.ViewToTarget(target), err)
}

func (s *TargetService) handleStopError(ctx context.Context, target *target.Target, err error) error {
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
