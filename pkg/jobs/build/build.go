// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"context"
	"errors"

	"github.com/daytonaio/daytona/pkg/build"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/telemetry"
)

type BuildJob struct {
	models.Job

	findBuild            func(ctx context.Context, buildId string) (*services.BuildDTO, error)
	listSuccessfulBuilds func(ctx context.Context, repoUrl string) ([]*models.Build, error)
	listConfigsForUrl    func(ctx context.Context, repoUrl string) ([]*models.GitProviderConfig, error)
	checkImageExists     func(ctx context.Context, image string) bool
	deleteImage          func(ctx context.Context, image string, force bool) error

	trackTelemetryEvent func(event telemetry.BuildRunnerEvent, clientId string, props map[string]interface{}) error
	loggerFactory       logs.ILoggerFactory
	builderFactory      build.IBuilderFactory

	basePath string
}

func (tj *BuildJob) Execute(ctx context.Context) error {
	switch tj.Action {
	case models.JobActionRun:
		return tj.run(ctx, &tj.Job)
	case models.JobActionDelete:
		return tj.delete(ctx, &tj.Job, false)
	case models.JobActionForceDelete:
		return tj.delete(ctx, &tj.Job, true)
	}
	return errors.New("invalid job action")
}
