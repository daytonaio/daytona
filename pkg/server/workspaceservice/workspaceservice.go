// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaceservice

import (
	"errors"

	p "github.com/daytonaio/daytona/pkg/server/provisioner"
	"github.com/daytonaio/daytona/pkg/types"
)

var provisioner = p.Provisioner{}

func getProject(workspace *types.Workspace, projectName string) (*types.Project, error) {
	for _, project := range workspace.Projects {
		if project.Name == projectName {
			return project, nil
		}
	}
	return nil, errors.New("project not found")
}
