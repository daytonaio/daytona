// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"errors"

	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/telemetry"
)

type WorkspaceJob struct {
	models.Job

	findWorkspace         func(ctx context.Context, workspaceId string) (*models.Workspace, error)
	findTarget            func(ctx context.Context, targetId string) (*models.Target, error)
	findContainerRegistry func(ctx context.Context, image string) (*models.ContainerRegistry, error)
	findGitProviderConfig func(ctx context.Context, id string) (*models.GitProviderConfig, error)
	trackTelemetryEvent   func(event telemetry.ServerEvent, clientId string, props map[string]interface{}) error

	loggerFactory logs.LoggerFactory
	provisioner   provisioner.IProvisioner
}

func (wj *WorkspaceJob) Execute(ctx context.Context) error {
	switch wj.Action {
	case models.JobActionCreate:
		return wj.create(ctx, &wj.Job)
	case models.JobActionStart:
		return wj.start(ctx, &wj.Job)
	case models.JobActionStop:
		return wj.stop(ctx, &wj.Job)
	case models.JobActionDelete:
		return wj.delete(ctx, &wj.Job, false)
	case models.JobActionForceDelete:
		return wj.delete(ctx, &wj.Job, true)
	}
	return errors.New("invalid job action")
}
