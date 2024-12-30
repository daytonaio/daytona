// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package os

import (
	"os"
)

func ChmodX(filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	err = f.Chmod(0755)
	if err != nil {
		return err
	}

	return nil
}
