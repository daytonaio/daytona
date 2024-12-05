// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/pkg/jobs/util"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/daytonaio/daytona/pkg/views"
)

func (wj *WorkspaceJob) start(ctx context.Context, j *models.Job) error {
	w, err := wj.findWorkspace(ctx, j.ResourceId)
	if err != nil {
		return err
	}

	workspaceLogger := wj.loggerFactory.CreateWorkspaceLogger(w.Id, w.Name, logs.LogSourceServer)
	defer workspaceLogger.Close()

	workspaceLogger.Write([]byte(fmt.Sprintf("Starting workspace %s\n", w.Name)))

	workspaceEnvVars, err := wj.getWorkspaceEnvironmentVariables(ctx, w)
	if err != nil {
		return err
	}
	w.EnvVars = workspaceEnvVars

	cr := wj.findContainerRegistry(ctx, w.Image, workspaceEnvVars)

	builderCr := wj.findContainerRegistry(ctx, wj.builderImage, workspaceEnvVars)

	var gc *models.GitProviderConfig

	if w.GitProviderConfigId != nil {
		gc, err = wj.findGitProviderConfig(ctx, *w.GitProviderConfigId)
		if err != nil && !stores.IsGitProviderNotFound(err) {
			return err
		}
	}

	w.EnvVars = util.ExtractContainerRegistryFromEnvVars(workspaceEnvVars)

	err = wj.provisioner.StartWorkspace(provisioner.WorkspaceParams{
		Workspace:                     w,
		ContainerRegistry:             cr,
		GitProviderConfig:             gc,
		BuilderImage:                  wj.builderImage,
		BuilderImageContainerRegistry: builderCr,
	})
	if err != nil {
		return err
	}

	workspaceLogger.Write([]byte(views.GetPrettyLogLine(fmt.Sprintf("Workspace %s started", w.Name))))
	return nil
}
