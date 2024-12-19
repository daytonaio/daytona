// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server/targets"
	"github.com/daytonaio/daytona/pkg/server/workspaces"
)

func HideDaytonaEnvVars(envVars *map[string]string) {
	for _, daytonaEnvVarKey := range getDaytonaEnvVarKeys() {
		delete(*envVars, daytonaEnvVarKey)
	}
}

func getDaytonaEnvVarKeys() []string {
	var result []string

	wsEnvVars := workspaces.GetWorkspaceEnvVars(&models.Workspace{
		Repository: &gitprovider.GitRepository{},
	}, workspaces.WorkspaceEnvVarParams{
		TelemetryEnabled: true,
	})

	targetEnvVars := targets.GetTargetEnvVars(&models.Target{}, targets.TargetEnvVarParams{})

	envVars := util.MergeEnvVars(wsEnvVars, targetEnvVars)

	for k := range envVars {
		result = append(result, k)
	}

	return result
}
