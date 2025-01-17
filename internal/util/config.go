// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import "os"

func DirectoryValidator(path *string) error {
	_, err := os.Stat(*path)
	if os.IsNotExist(err) {
		return os.MkdirAll(*path, 0700)
	}
	return err
}
