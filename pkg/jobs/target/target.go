// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"
	"errors"

	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/runner/providermanager"
	"github.com/daytonaio/daytona/pkg/telemetry"
)

type TargetJob struct {
	models.Job

	findTarget               func(ctx context.Context, targetId string) (*models.Target, error)
	handleSuccessfulCreation func(ctx context.Context, targetId string) error

	trackTelemetryEvent          func(event telemetry.ServerEvent, clientId string, props map[string]interface{}) error
	updateTargetProviderMetadata func(ctx context.Context, targetId, metadata string) error

	loggerFactory   logs.ILoggerFactory
	providerManager providermanager.IProviderManager
}

func (tj *TargetJob) Execute(ctx context.Context) error {
	switch tj.Action {
	case models.JobActionCreate:
		return tj.create(ctx, &tj.Job)
	case models.JobActionStart:
		return tj.start(ctx, &tj.Job)
	case models.JobActionStop:
		return tj.stop(ctx, &tj.Job)
	case models.JobActionRestart:
		return tj.restart(ctx, &tj.Job)
	case models.JobActionDelete:
		return tj.delete(ctx, &tj.Job, false)
	case models.JobActionForceDelete:
		return tj.delete(ctx, &tj.Job, true)
	}
	return errors.New("invalid job action")
}
