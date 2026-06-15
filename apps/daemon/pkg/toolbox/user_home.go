//go:build !windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package toolbox

import "os"

func getUserHomeDir() (string, error) {
	return os.UserHomeDir()
}
