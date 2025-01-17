// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import "github.com/daytonaio/daytona/pkg/models"

type TargetEnvVarParams struct {
	ApiUrl           string
	ServerUrl        string
	ServerVersion    string
	ClientId         string
	TelemetryEnabled bool
}

func GetTargetEnvVars(target *models.Target, params TargetEnvVarParams) map[string]string {
	envVars := map[string]string{
		"DAYTONA_TARGET_ID":      target.Id,
		"DAYTONA_SERVER_API_KEY": target.ApiKey,
		"DAYTONA_SERVER_VERSION": params.ServerVersion,
		"DAYTONA_SERVER_URL":     params.ServerUrl,
		"DAYTONA_SERVER_API_URL": params.ApiUrl,
		"DAYTONA_CLIENT_ID":      params.ClientId,
		// (HOME) will be replaced at runtime
		"DAYTONA_AGENT_LOG_FILE_PATH": "(HOME)/.daytona-agent.log",
	}

	if params.TelemetryEnabled {
		envVars["DAYTONA_TELEMETRY_ENABLED"] = "true"
	}

	return envVars
}
