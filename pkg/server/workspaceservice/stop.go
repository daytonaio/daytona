// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaceservice

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/server/db"
	"github.com/daytonaio/daytona/pkg/server/targets"

	log "github.com/sirupsen/logrus"
)

func StopWorkspace(workspaceId string) error {
	workspace, err := db.FindWorkspaceByIdOrName(workspaceId)
	if err != nil {
		return errors.New("workspace not found")
	}

	log.Info("Stopping workspace")

	target, err := targets.GetTarget(workspace.Target)
	if err != nil {
		return err
	}

	for _, project := range workspace.Projects {
		//	todo: go routines
		err := provisioner.StopProject(project, target)
		if err != nil {
			return err
		}
	}

	err = provisioner.StopWorkspace(workspace, target)
	if err != nil {
		return err
	}

	log.Info("Workspace stopped")
	return nil
}

func StopProject(workspaceId, projectId string) error {
	w, err := db.FindWorkspaceByIdOrName(workspaceId)
	if err != nil {
		return errors.New("workspace not found")
	}

	project, err := getProject(w, projectId)
	if err != nil {
		return errors.New("project not found")
	}

	target, err := targets.GetTarget(project.Target)
	if err != nil {
		return err
	}

	return provisioner.StopProject(project, target)
}
