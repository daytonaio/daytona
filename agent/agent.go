// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	log "github.com/sirupsen/logrus"
)

func Start() error {
	log.Info("Starting Daytona Agent")

	log.Info(config)

	return nil
}
