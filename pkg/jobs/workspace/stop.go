// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/views"
)

func (wj *WorkspaceJob) stop(ctx context.Context, j *models.Job) error {
	w, err := wj.findWorkspace(ctx, j.ResourceId)
	if err != nil {
		return err
	}

	workspaceLogger, err := wj.loggerFactory.CreateWorkspaceLogger(w.Id, w.Name, logs.LogSourceServer)
	if err != nil {
		return err
	}
	defer workspaceLogger.Close()

	workspaceLogger.Write([]byte(fmt.Sprintf("Stopping workspace %s\n", w.Name)))

	p, err := wj.providerManager.GetProvider(w.Target.TargetConfig.ProviderInfo.Name)
	if err != nil {
		return err
	}

	_, err = (*p).StopWorkspace(&provider.WorkspaceRequest{
		Workspace: w,
	})
	if err != nil {
		return err
	}

	workspaceLogger.Write([]byte(views.GetPrettyLogLine(fmt.Sprintf("Workspace %s stopped", w.Name))))

	return nil
}
