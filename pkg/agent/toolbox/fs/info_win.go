//go:build windows

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package fs

import (
	"fmt"
	"os"
	"strconv"

	"golang.org/x/sys/windows"
)

func getFileInfo(path string) (*FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return &FileInfo{}, err
	}

	ownerSid, groupSid, err := getFileGidUid(path)
	if err != nil {
		return &FileInfo{}, err
	}

	return &FileInfo{
		Name:        info.Name(),
		Size:        info.Size(),
		Mode:        info.Mode().String(),
		ModTime:     info.ModTime().String(),
		IsDir:       info.IsDir(),
		Owner:       ownerSid,
		Group:       groupSid,
		Permissions: fmt.Sprintf("%04o", info.Mode().Perm()),
	}, nil
}

func getFileGidUid(path string) (string, string, error) {
	sd, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.GROUP_SECURITY_INFORMATION)
	if err != nil {
		return "", "", err
	}
	owner, _, err := sd.Owner()
	if err != nil {
		return "", "", err
	}
	group, _, err := sd.Group()
	if err != nil {
		return "", "", err
	}
	return owner.String(), group.String(), nil
}

func GetFileUid(stat os.FileInfo) (uint32, error) {
	uid, _, err := getFileGidUid(stat.Name())
	if err != nil {
		return 0, err
	}
	uidInt, err := strconv.ParseUint(uid, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(uidInt), nil
}
