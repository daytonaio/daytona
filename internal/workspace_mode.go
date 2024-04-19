// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package internal

import "os"

func WorkspaceMode() bool {
	_, devEnv := os.LookupEnv("DAYTONA_DEV")
	if devEnv {
		return false
	}
	val, wsMode := os.LookupEnv("DAYTONA_WS_ID")
	if wsMode && val != "" {
		return true
	}
	return false
}
