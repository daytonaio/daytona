// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
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

	cr, err := wj.findContainerRegistry(ctx, w.Image)
	if err != nil && !stores.IsContainerRegistryNotFound(err) {
		return err
	}

	var gc *models.GitProviderConfig

	if w.GitProviderConfigId != nil {
		gc, err = wj.findGitProviderConfig(ctx, *w.GitProviderConfigId)
		if err != nil && !stores.IsGitProviderNotFound(err) {
			return err
		}
	}

	err = wj.provisioner.CreateWorkspace(w, cr, gc)
	if err != nil {
		return err
	}

	workspaceLogger.Write([]byte(views.GetPrettyLogLine(fmt.Sprintf("Workspace %s created", w.Name))))
	return wj.start(ctx, j)
}
