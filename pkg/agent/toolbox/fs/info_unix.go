//go:build !windows

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package fs

import (
	"fmt"
	"os"
	"strconv"
	"syscall"
)

func getFileInfo(path string) (*FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return &FileInfo{}, err
	}

	stat := info.Sys().(*syscall.Stat_t)
	return &FileInfo{
		Name:        info.Name(),
		Size:        info.Size(),
		Mode:        info.Mode().String(),
		ModTime:     info.ModTime().String(),
		IsDir:       info.IsDir(),
		Owner:       strconv.FormatUint(uint64(stat.Uid), 10),
		Group:       strconv.FormatUint(uint64(stat.Gid), 10),
		Permissions: fmt.Sprintf("%04o", info.Mode().Perm()),
	}, nil
}

func GetFileUid(stat os.FileInfo) (uint32, error) {
	return stat.Sys().(*syscall.Stat_t).Uid, nil
}
