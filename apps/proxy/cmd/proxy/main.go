// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package main

import (
	"github.com/daytonaio/proxy/cmd/proxy/config"
	"github.com/daytonaio/proxy/pkg/proxy"

	log "github.com/sirupsen/logrus"
)

func main() {
	config, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	err = proxy.StartProxy(config)
	if err != nil {
		log.Fatal(err)
	}
}
