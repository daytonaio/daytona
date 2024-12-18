// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/provider"
	log "github.com/sirupsen/logrus"
)

func (wj *WorkspaceJob) delete(ctx context.Context, j *models.Job, force bool) error {
	w, err := wj.findWorkspace(ctx, j.ResourceId)
	if err != nil {
		return err
	}

	workspaceLogger, err := wj.loggerFactory.CreateLogger(w.Id, w.Name, logs.LogSourceServer)
	if err != nil {
		return err
	}
	defer workspaceLogger.Close()

	workspaceLogger.Write([]byte(fmt.Sprintf("Destroying workspace %s", w.Name)))

	p, err := wj.providerManager.GetProvider(w.Target.TargetConfig.ProviderInfo.Name)
	if err != nil {
		if force {
			log.Error(err)
			return nil
		}
		return err
	}

	_, err = (*p).DestroyWorkspace(&provider.WorkspaceRequest{
		Workspace: w,
	})
	if err != nil {
		if force {
			log.Error(err)
		}
		return nil
	}

	workspaceLogger.Write([]byte(fmt.Sprintf("Workspace %s destroyed", w.Name)))

	err = workspaceLogger.Cleanup()
	if err != nil {
		// Should not fail the whole operation if the workspace logger cannot be cleaned up
		log.Error(err)
	}

	return nil
}
