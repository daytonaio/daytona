// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaceservice

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/server/db"
	"github.com/daytonaio/daytona/pkg/server/provisioner"
)

func StartWorkspace(workspaceId string) error {
	w, err := db.FindWorkspaceByIdOrName(workspaceId)
	if err != nil {
		return errors.New("workspace not found")
	}

	return provisioner.StartWorkspace(w)
}

func StartProject(workspaceId, projectId string) error {
	w, err := db.FindWorkspaceByIdOrName(workspaceId)
	if err != nil {
		return errors.New("workspace not found")
	}

	project, err := getProject(w, projectId)
	if err != nil {
		return errors.New("project not found")
	}

	return provisioner.StartProject(project)
}
