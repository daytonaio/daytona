//go:build linux

// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package fs

import (
	"os"
	"strconv"
	"syscall"
)

func ownerGroup(info os.FileInfo) (owner, group string) {
	stat := info.Sys().(*syscall.Stat_t)
	return strconv.FormatUint(uint64(stat.Uid), 10), strconv.FormatUint(uint64(stat.Gid), 10)
}
