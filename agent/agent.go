// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"github.com/daytonaio/daytona/agent/config"
	"github.com/daytonaio/daytona/agent/tailscale"
	log "github.com/sirupsen/logrus"
)

func Start() error {
	log.Info("Starting Daytona Agent")

	c, err := config.GetConfig()
	if err != nil {
		return err
	}

	tailscale.Start(c)

	return nil
}
