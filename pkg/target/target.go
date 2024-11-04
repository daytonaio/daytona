// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

type Target struct {
	Id           string            `json:"id" validate:"required"`
	Name         string            `json:"name" validate:"required"`
	TargetConfig string            `json:"targetConfig" validate:"required"`
	ApiKey       string            `json:"-"`
	EnvVars      map[string]string `json:"-"`
} // @name Target

type TargetInfo struct {
	Name             string `json:"name" validate:"required"`
	ProviderMetadata string `json:"providerMetadata,omitempty" validate:"optional"`
} // @name TargetInfo

type TargetEnvVarParams struct {
	ApiUrl        string
	ServerUrl     string
	ServerVersion string
	ClientId      string
}

func GetTargetEnvVars(target *Target, params TargetEnvVarParams, telemetryEnabled bool) map[string]string {
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

	if telemetryEnabled {
		envVars["DAYTONA_TELEMETRY_ENABLED"] = "true"
	}

	return envVars
}
