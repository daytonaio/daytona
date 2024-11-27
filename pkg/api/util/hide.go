// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"encoding/json"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/server/targets"
	"github.com/daytonaio/daytona/pkg/server/workspaces"
)

func GetMaskedOptions(server *server.Server, providerName, options string) (string, error) {
	p, err := server.ProviderManager.GetProvider(providerName)
	if err != nil {
		return "", err
	}

	manifest, err := (*p).GetTargetConfigManifest()
	if err != nil {
		return "", err
	}

	var opts map[string]interface{}
	err = json.Unmarshal([]byte(options), &opts)
	if err != nil {
		return "", err
	}

	for name, property := range *manifest {
		if property.InputMasked {
			delete(opts, name)
		}
	}

	updatedOptions, err := json.MarshalIndent(opts, "", "  ")
	if err != nil {
		return "", err
	}

	return string(updatedOptions), nil
}

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
