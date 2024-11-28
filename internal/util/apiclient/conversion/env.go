// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package conversion

import "github.com/daytonaio/daytona/pkg/apiclient"

func ToEnvVarsMap(envVars []apiclient.EnvironmentVariable) map[string]string {
	envVarsMap := map[string]string{}

	for _, envVar := range envVars {
		envVarsMap[envVar.Key] = envVar.Value
	}

	return envVarsMap
}
