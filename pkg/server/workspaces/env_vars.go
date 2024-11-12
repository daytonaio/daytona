// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import "github.com/daytonaio/daytona/pkg/models"

type WorkspaceEnvVarParams struct {
	ApiUrl        string
	ServerUrl     string
	ServerVersion string
	ClientId      string
}

func GetWorkspaceEnvVars(workspace *models.Workspace, params WorkspaceEnvVarParams, telemetryEnabled bool) map[string]string {
	envVars := map[string]string{
		"DAYTONA_TARGET_ID":                workspace.TargetId,
		"DAYTONA_WORKSPACE_ID":             workspace.Id,
		"DAYTONA_WORKSPACE_REPOSITORY_URL": workspace.Repository.Url,
		"DAYTONA_SERVER_API_KEY":           workspace.ApiKey,
		"DAYTONA_SERVER_VERSION":           params.ServerVersion,
		"DAYTONA_SERVER_URL":               params.ServerUrl,
		"DAYTONA_SERVER_API_URL":           params.ApiUrl,
		"DAYTONA_CLIENT_ID":                params.ClientId,
		// (HOME) will be replaced at runtime
		"DAYTONA_AGENT_LOG_FILE_PATH": "(HOME)/.daytona-agent.log",
	}

	if telemetryEnabled {
		envVars["DAYTONA_TELEMETRY_ENABLED"] = "true"
	}

	return envVars
}
