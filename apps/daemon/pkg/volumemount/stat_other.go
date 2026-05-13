// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build !linux

package volumemount

import "os"

// statDev is a no-op on non-Linux platforms. The daemon only runs inside
// sandbox containers (Linux), but the package still needs to compile for dev
// environments (e.g. macOS) so engineers can build/test it.
func statDev(_ os.FileInfo) (uint64, bool) {
	return 0, false
}
