// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import "github.com/daytonaio/daytona/pkg/apiclient"

func IsProjectRunning(workspace *apiclient.WorkspaceDTO, projectName string) bool {
	for _, project := range workspace.GetProjects() {
		if project.GetName() == projectName {
			uptime := project.GetState().Uptime
			return uptime != nil && *uptime != 0
		}
	}
	return false
}
