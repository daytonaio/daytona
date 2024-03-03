// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package os

import "os/exec"

func ChmodX(filePath string) error {
	err := exec.Command("chmod", "+x", filePath).Run()
	if err != nil {
		return err
	}

	return nil
}
