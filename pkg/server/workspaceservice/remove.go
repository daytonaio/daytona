// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaceservice

import (
	"errors"
	"fmt"

	"github.com/daytonaio/daytona/pkg/server/auth"
	"github.com/daytonaio/daytona/pkg/server/config"
	"github.com/daytonaio/daytona/pkg/server/db"
	"github.com/daytonaio/daytona/pkg/server/targets"

	log "github.com/sirupsen/logrus"
)

func RemoveWorkspace(workspaceId string) error {
	workspace, err := db.FindWorkspaceByIdOrName(workspaceId)
	if err != nil {
		return errors.New("workspace not found")
	}

	log.Infof("Destroying workspace %s", workspace.Id)

	target, err := targets.GetTarget(workspace.Target)
	if err != nil {
		return err
	}

	for _, project := range workspace.Projects {
		//	todo: go routines
		err := provisioner.DestroyProject(project, target)
		if err != nil {
			return err
		}
	}

	err = provisioner.DestroyWorkspace(workspace, target)
	if err != nil {
		return err
	}

	err = config.DeleteWorkspaceLogs(workspace.Id)
	if err != nil {
		return err
	}

	for _, project := range workspace.Projects {
		err := auth.RevokeApiKey(fmt.Sprintf("%s/%s", workspace.Id, project.Name))
		if err != nil {
			// Should not fail the whole operation if the API key cannot be revoked
			log.Error(err)
		}
	}

	err = db.DeleteWorkspace(workspace)
	if err != nil {
		return err
	}

	log.Infof("Workspace %s destroyed", workspace.Id)
	return nil
}
