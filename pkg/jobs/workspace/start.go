// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/daytonaio/daytona/pkg/views"
)

func (wj *WorkspaceJob) start(ctx context.Context, j *models.Job) error {
	w, err := wj.findWorkspace(ctx, j.ResourceId)
	if err != nil {
		return err
	}

	workspaceLogger, err := wj.loggerFactory.CreateWorkspaceLogger(w.Id, w.Name, logs.LogSourceServer)
	if err != nil {
		return err
	}
	defer workspaceLogger.Close()

	workspaceLogger.Write([]byte(fmt.Sprintf("Starting workspace %s\n", w.Name)))

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

	p, err := wj.providerManager.GetProvider(w.Target.TargetConfig.ProviderInfo.Name)
	if err != nil {
		return err
	}

	req := &provider.WorkspaceRequest{
		Workspace:           w,
		ContainerRegistries: containerRegistries,
		GitProviderConfig:   gc,
		BuilderImage:        wj.builderImage,
	}

	_, err = (*p).StartWorkspace(req)
	if err != nil {
		return err
	}

	providerMetadata, err := (*p).GetWorkspaceProviderMetadata(req)
	if err != nil {
		return err
	}

	err = wj.updateWorkspaceProviderMetadata(ctx, w.Id, providerMetadata)
	if err != nil {
		return err
	}

	workspaceLogger.Write([]byte(views.GetPrettyLogLine(fmt.Sprintf("Workspace %s started", w.Name))))
	return nil
}
