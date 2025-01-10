// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package common

import "os"

func AgentMode() bool {
	_, devEnv := os.LookupEnv("DAYTONA_DEV")
	if devEnv {
		return false
	}
	val, agentMode := os.LookupEnv("DAYTONA_TARGET_ID")
	if agentMode && val != "" {
		return true
	}
	return false
}
