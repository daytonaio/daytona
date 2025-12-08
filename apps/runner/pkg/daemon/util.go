// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package daemon

import (
	"fmt"
	"os"
	"path/filepath"
)

func WriteStaticBinary(name string) (string, error) {
	daemonBinary, err := static.ReadFile(fmt.Sprintf("static/%s", name))
	if err != nil {
		return "", err
	}

	// Use the mounted binaries directory for Kubernetes/containerd deployments
	// This ensures binaries are on the node filesystem for containerd to mount them into containers
	binariesDir := os.Getenv("DAEMON_BINARIES_DIR")
	if binariesDir == "" {
		// Fallback for local development
		pwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		binariesDir = filepath.Join(pwd, ".tmp", "binaries")
	}

	err = os.MkdirAll(binariesDir, 0755)
	if err != nil {
		return "", err
	}

	daemonPath := filepath.Join(binariesDir, name)
	_, err = os.Stat(daemonPath)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	if err == nil {
		err = os.Remove(daemonPath)
		if err != nil {
			return "", err
		}
	}

	err = os.WriteFile(daemonPath, daemonBinary, 0755)
	if err != nil {
		return "", err
	}

	return daemonPath, nil
}
