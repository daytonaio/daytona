// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"strings"
)

func IsLocalDockerTarget(providerName, options, runnerId string) bool {
	if providerName != "docker-provider" {
		return false
	}

	return !strings.Contains(options, "Remote Hostname") && runnerId == LOCAL_RUNNER_ID
}
