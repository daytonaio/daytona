// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/daytonaio/daytona/pkg/views"
)

func (wj *WorkspaceJob) create(ctx context.Context, j *models.Job) error {
	w, err := wj.findWorkspace(ctx, j.ResourceId)
	if err != nil {
		return err
	}

	workspaceLogger := wj.loggerFactory.CreateWorkspaceLogger(w.Id, w.Name, logs.LogSourceServer)
	defer workspaceLogger.Close()

	workspaceLogger.Write([]byte(fmt.Sprintf("Creating workspace %s\n", w.Name)))

	workspaceEnvVars, err := wj.getWorkspaceEnvironmentVariables(ctx, w)
	if err != nil {
		return err
	}

	var gc *models.GitProviderConfig

	if w.GitProviderConfigId != nil {
		gc, err = wj.findGitProviderConfig(ctx, *w.GitProviderConfigId)
		if err != nil && !stores.IsGitProviderNotFound(err) {
			return err
		}
	}

	extractedEnvVars, containerRegistries := common.ExtractContainerRegistryFromEnvVars(workspaceEnvVars)

	w.EnvVars = extractedEnvVars

	err = wj.provisioner.CreateWorkspace(provisioner.WorkspaceParams{
		Workspace:           w,
		ContainerRegistries: containerRegistries,
		GitProviderConfig:   gc,
		BuilderImage:        wj.builderImage,
	})
	if err != nil {
		return err
	}

	workspaceLogger.Write([]byte(views.GetPrettyLogLine(fmt.Sprintf("Workspace %s created", w.Name))))
	return wj.start(ctx, j)
}
