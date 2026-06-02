// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build !linux

package volumemount

import "os"

// statDev is a no-op on non-Linux platforms; it exists only so the package
// builds in dev environments (e.g. macOS). The daemon runs on Linux.
func statDev(_ os.FileInfo) (uint64, bool) {
	return 0, false
}
