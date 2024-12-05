// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import "strings"

func ExtractContainerRegistryFromEnvVars(envVars map[string]string) map[string]string {
	result := make(map[string]string)

	for k, v := range envVars {
		if !strings.HasSuffix(k, "CONTAINER_REGISTRY_SERVER") && !strings.HasSuffix(k, "CONTAINER_REGISTRY_USERNAME") && !strings.HasSuffix(k, "CONTAINER_REGISTRY_PASSWORD") {
			result[k] = v
		}
	}

	return result
}
