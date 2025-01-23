// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"context"
	"errors"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/runner/providermanager"
	"github.com/daytonaio/daytona/pkg/telemetry"
)

type RunnerJob struct {
	models.Job

	trackTelemetryEvent func(event telemetry.Event, clientId string) error
	providerManager     providermanager.IProviderManager
}

func (pj *RunnerJob) Execute(ctx context.Context) error {
	switch pj.Action {
	case models.JobActionInstallProvider:
		return pj.installProvider(ctx, &pj.Job)
	case models.JobActionUpdateProvider:
		return pj.updateProvider(ctx, &pj.Job)
	case models.JobActionUninstallProvider:
		return pj.uninstallProvider(ctx, &pj.Job)
	}
	return errors.New("invalid job action")
}
