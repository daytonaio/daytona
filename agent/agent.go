// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"github.com/daytonaio/daytona/agent/tailscale"
	log "github.com/sirupsen/logrus"
)

func Start() error {
	log.Info("Starting Daytona Agent")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	tailscale.Start(config.ReverseProxy.AuthKey)

	return nil
}
