// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"context"
	"fmt"
	"io"

	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/telemetry"
	log "github.com/sirupsen/logrus"

	"github.com/daytonaio/daytona/internal/util"
)

func (s *TargetService) StartTarget(ctx context.Context, targetId string) error {
	t, err := s.targetStore.Find(targetId)
	if err != nil {
		return s.handleStartError(ctx, nil, ErrTargetNotFound)
	}

	targetConfig, err := s.targetConfigStore.Find(&provider.TargetConfigFilter{Name: &t.TargetConfig})
	if err != nil {
		return s.handleStartError(ctx, t, err)
	}

	targetLogger := s.loggerFactory.CreateTargetLogger(t.Id, logs.LogSourceServer)
	defer targetLogger.Close()

	logger := io.MultiWriter(&util.InfoLogWriter{}, targetLogger)

	t.EnvVars = target.GetTargetEnvVars(t, target.TargetEnvVarParams{
		ApiUrl:        s.serverApiUrl,
		ServerUrl:     s.serverUrl,
		ServerVersion: s.serverVersion,
		ClientId:      telemetry.ClientId(ctx),
	}, telemetry.TelemetryEnabled(ctx))

	err = s.startTarget(t, targetConfig, logger)
	if err != nil {
		return s.handleStartError(ctx, t, err)
	}

	return s.handleStartError(ctx, t, err)
}

func (s *TargetService) startTarget(target *target.Target, targetConfig *provider.TargetConfig, targetLogger io.Writer) error {
	targetLogger.Write([]byte("Starting target\n"))

	err := s.provisioner.StartTarget(target, targetConfig)
	if err != nil {
		return err
	}

	targetLogger.Write([]byte(fmt.Sprintf("Target %s started\n", target.Name)))

	return err
}

func (s *TargetService) handleStartError(ctx context.Context, target *target.Target, err error) error {
	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	clientId := telemetry.ClientId(ctx)

	telemetryProps := telemetry.NewTargetEventProps(ctx, target, nil)
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
