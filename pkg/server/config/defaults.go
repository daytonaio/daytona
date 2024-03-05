// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/types"
)

const defaultRegistryUrl = "https://download.daytona.io/daytona"
const defaultServerDownloadUrl = "https://download.daytona.io/daytona/get-server.sh"
const defaultHeadscalePort = 3001
const defaultApiPort = 3000

var us_defaultFrpsConfig = types.FRPSConfig{
	Domain:   "try-us.daytona.app",
	Port:     7000,
	Protocol: "https",
}

var eu_defaultFrpsConfig = types.FRPSConfig{
	Domain:   "try-eu.daytona.app",
	Port:     7000,
	Protocol: "https",
}

func getDefaultFRPSConfig() *types.FRPSConfig {
	// Return config which responds fastest to a ping
	usReturnChan := make(chan bool)
	euReturnChan := make(chan bool)

	go func() {
		// Ping US server
		_, _ = http.Get(fmt.Sprintf("%s://%s:%d", us_defaultFrpsConfig.Protocol, us_defaultFrpsConfig.Domain, us_defaultFrpsConfig.Port))
		usReturnChan <- true
	}()

	go func() {
		// Ping EU server
		_, _ = http.Get(fmt.Sprintf("%s://%s:%d", eu_defaultFrpsConfig.Protocol, eu_defaultFrpsConfig.Domain, eu_defaultFrpsConfig.Port))
		euReturnChan <- true
	}()

	select {
	case <-usReturnChan:
		return &us_defaultFrpsConfig
	case <-euReturnChan:
		return &eu_defaultFrpsConfig
	}
}
