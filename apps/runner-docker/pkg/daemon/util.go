// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package daemon

import (
	"os"
	"path/filepath"
)

func WriteDaemonBinary() (string, error) {
	daemonBinary, err := static.ReadFile("static/daemon-amd64")
	if err != nil {
		return "", err
	}

	tmpDir := os.TempDir()
	daemonPath := filepath.Join(tmpDir, "daemon-amd64")

	_, err = os.Stat(daemonPath)
	if err == nil {
		err = os.Remove(daemonPath)
		if err != nil {
			return "", err
		}
	} else if !os.IsNotExist(err) {
		return "", err
	}

	err = os.WriteFile(daemonPath, daemonBinary, 0755)
	if err != nil {
		return "", err
	}

	return daemonPath, nil
}
