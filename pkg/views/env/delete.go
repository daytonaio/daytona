// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package env

import (
	"context"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views/env/selection"
)

func RemoveEnvVarsView(ctx context.Context, apiClient apiclient.APIClient) ([]*apiclient.EnvironmentVariable, error) {
	envVars, res, err := apiClient.EnvVarAPI.ListEnvironmentVariables(ctx).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	selectedEnvVars := selection.GetEnvironmentVariablesFromPrompt(envVars, "Remove")

	return selectedEnvVars, nil
}
