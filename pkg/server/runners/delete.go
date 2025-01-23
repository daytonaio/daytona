// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runners

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/daytonaio/daytona/pkg/telemetry"
	log "github.com/sirupsen/logrus"
)

func (s *RunnerService) Delete(ctx context.Context, runnerId string) error {
	var err error
	ctx, err = s.runnerStore.BeginTransaction(ctx)
	if err != nil {
		return s.handleDeleteError(ctx, nil, err)
	}

	defer stores.RecoverAndRollback(ctx, s.runnerStore)

	runner, err := s.runnerStore.Find(ctx, runnerId)
	if err != nil {
		return s.handleDeleteError(ctx, nil, err)
	}

	err = s.runnerStore.Delete(ctx, runner)
	if err != nil {
		return s.handleDeleteError(ctx, runner, err)
	}

	metadata, err := s.runnerMetadataStore.Find(ctx, runnerId)
	if err != nil && !stores.IsRunnerMetadataNotFound(err) {
		return s.handleDeleteError(ctx, runner, err)
	}
	if metadata != nil {
		err = s.runnerMetadataStore.Delete(ctx, metadata)
		if err != nil {
			return s.handleDeleteError(ctx, runner, err)
		}
	}

	err = s.deleteApiKey(ctx, runner.Id)
	if err != nil {
		return s.handleDeleteError(ctx, runner, err)
	}

	err = s.unsetDefaultTarget(ctx, runner.Id)
	if err != nil {
		return s.handleDeleteError(ctx, runner, err)
	}

	err = s.runnerStore.CommitTransaction(ctx)
	return s.handleDeleteError(ctx, runner, err)
}

func (s *RunnerService) handleDeleteError(ctx context.Context, r *models.Runner, err error) error {
	if err != nil {
		err = s.runnerStore.RollbackTransaction(ctx, err)
	}

	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	clientId := telemetry.ClientId(ctx)

	eventName := telemetry.RunnerEventLifecycleDeleted
	if err != nil {
		eventName = telemetry.RunnerEventLifecycleDeletionFailed
	}
	event := telemetry.NewRunnerEvent(eventName, r, err, nil)

	telemetryError := s.trackTelemetryEvent(event, clientId)
	if telemetryError != nil {
		log.Trace(telemetryError)
	}

	return err
}
