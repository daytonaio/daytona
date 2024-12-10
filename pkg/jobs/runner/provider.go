// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"context"
	"errors"

	"github.com/daytonaio/daytona/pkg/build"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/telemetry"
)

type RunnerJob struct {
	models.Job

	findBuild            func(ctx context.Context, buildId string) (*services.BuildDTO, error)
	listSuccessfulBuilds func(ctx context.Context, repoUrl string) ([]*models.Build, error)
	checkImageExists     func(ctx context.Context, image string) bool
	deleteImage          func(ctx context.Context, image string, force bool) error

	trackTelemetryEvent func(event telemetry.BuildRunnerEvent, clientId string, props map[string]interface{}) error
	loggerFactory       logs.LoggerFactory
	builderFactory      build.IBuilderFactory

	basePath string
}

func (pj *RunnerJob) Execute(ctx context.Context) error {
	switch pj.Action {
	case models.JobActionInstallProvider:
		return pj.install(ctx, &pj.Job)
	case models.JobActionUpdateProvider:
		return pj.update(ctx, &pj.Job)
	case models.JobActionUninstallProvider:
		return pj.uninstall(ctx, &pj.Job)
	}
	return errors.New("invalid job action")
}
