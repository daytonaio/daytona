// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runners

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/daytonaio/daytona/pkg/telemetry"

	log "github.com/sirupsen/logrus"
)

func (s *RunnerService) CreateRunner(ctx context.Context, req services.CreateRunnerDTO) (*services.RunnerDTO, error) {
	var err error
	ctx, err = s.runnerStore.BeginTransaction(ctx)
	if err != nil {
		return nil, s.handleRegisterError(ctx, nil, err)
	}

	defer stores.RecoverAndRollback(ctx, s.runnerStore)

	_, err = s.runnerStore.Find(ctx, req.Name)
	if err == nil {
		return nil, s.handleRegisterError(ctx, nil, services.ErrRunnerAlreadyExists)
	}

	apiKey, err := s.createApiKey(ctx, req.Id)
	if err != nil {
		return nil, s.handleRegisterError(ctx, nil, err)
	}

	runner := &models.Runner{
		Id:     req.Id,
		Name:   req.Name,
		ApiKey: apiKey,
		Metadata: &models.RunnerMetadata{
			RunnerId: req.Id,
			Uptime:   0,
		},
	}

	err = s.runnerStore.Save(ctx, runner)
	if err != nil {
		return nil, s.handleRegisterError(ctx, runner, err)
	}

	err = s.runnerMetadataStore.Save(ctx, runner.Metadata)
	if err != nil {
		return nil, s.handleRegisterError(ctx, runner, err)
	}

	err = s.runnerStore.CommitTransaction(ctx)
	if err != nil {
		return nil, s.handleRegisterError(ctx, runner, err)
	}

	return &services.RunnerDTO{
		Runner: *runner,
		State:  runner.GetState(),
	}, s.handleRegisterError(ctx, runner, nil)
}

func (s *RunnerService) handleRegisterError(ctx context.Context, r *models.Runner, err error) error {
	if err != nil {
		err = s.runnerStore.RollbackTransaction(ctx, err)
	}

	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	clientId := telemetry.ClientId(ctx)

	eventName := telemetry.RunnerEventLifecycleRegistered
	if err != nil {
		eventName = telemetry.RunnerEventLifecycleRegistrationFailed
	}
	event := telemetry.NewRunnerEvent(eventName, r, err, nil)

	telemetryError := s.trackTelemetryEvent(event, clientId)
	if telemetryError != nil {
		log.Trace(telemetryError)
	}

	return err
}
