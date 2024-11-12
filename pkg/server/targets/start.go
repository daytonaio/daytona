// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"context"
	"fmt"
	"io"

	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/daytonaio/daytona/pkg/views"
	log "github.com/sirupsen/logrus"

	"github.com/daytonaio/daytona/internal/util"
)

func (s *TargetService) StartTarget(ctx context.Context, targetId string) error {
	t, err := s.targetStore.Find(&TargetFilter{IdOrName: &targetId})
	if err != nil {
		return s.handleStartError(ctx, nil, ErrTargetNotFound)
	}

	targetLogger := s.loggerFactory.CreateTargetLogger(t.Id, t.Name, logs.LogSourceServer)
	defer targetLogger.Close()

	logger := io.MultiWriter(&util.InfoLogWriter{}, targetLogger)

	t.EnvVars = GetTargetEnvVars(t, TargetEnvVarParams{
		ApiUrl:        s.serverApiUrl,
		ServerUrl:     s.serverUrl,
		ServerVersion: s.serverVersion,
		ClientId:      telemetry.ClientId(ctx),
	}, telemetry.TelemetryEnabled(ctx))

	err = s.startTarget(t, logger)
	if err != nil {
		return s.handleStartError(ctx, t, err)
	}

	return s.handleStartError(ctx, t, err)
}

func (s *TargetService) startTarget(target *models.Target, targetLogger io.Writer) error {
	targetLogger.Write([]byte("Starting target\n"))

	err := s.provisioner.StartTarget(target)
	if err != nil {
		return err
	}

	targetLogger.Write([]byte(views.GetPrettyLogLine(fmt.Sprintf("Target %s started", target.Name))))

	return err
}

func (s *TargetService) handleStartError(ctx context.Context, target *models.Target, err error) error {
	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	clientId := telemetry.ClientId(ctx)

	telemetryProps := telemetry.NewTargetEventProps(ctx, target)
	event := telemetry.ServerEventTargetStarted
	if err != nil {
		telemetryProps["error"] = err.Error()
		event = telemetry.ServerEventTargetStartError
	}
	telemetryError := s.telemetryService.TrackServerEvent(event, clientId, telemetryProps)
	if telemetryError != nil {
		log.Trace(telemetryError)
	}

	return err
}
