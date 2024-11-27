// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	log "github.com/sirupsen/logrus"
)

func (wj *WorkspaceJob) delete(ctx context.Context, j *models.Job, force bool) error {
	w, err := wj.findWorkspace(ctx, j.ResourceId)
	if err != nil {
		return err
	}

	workspaceLogger := wj.loggerFactory.CreateWorkspaceLogger(w.Id, w.Name, logs.LogSourceServer)

	workspaceLogger.Write([]byte(fmt.Sprintf("Destroying workspace %s", w.Name)))

	err = wj.provisioner.DestroyWorkspace(w)
	if err != nil {
		if !force {
			return err
		}
		log.Error(err)
	}

	workspaceLogger.Write([]byte(fmt.Sprintf("Workspace %s destroyed", w.Name)))

	err = workspaceLogger.Cleanup()
	if err != nil {
		// Should not fail the whole operation if the workspace logger cannot be cleaned up
		log.Error(err)
	}

	return nil
}
