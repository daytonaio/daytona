// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"path/filepath"

	"github.com/daytonaio/daytona/cmd/daytona/config"
)

func GetBuildLogsDir() (string, error) {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "builds", "logs"), nil
}
