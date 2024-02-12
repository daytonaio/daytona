// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	log "github.com/sirupsen/logrus"
)

type Self struct {
	HostName string `json:"host_name"`
	DNSName  string `json:"dns_name"`
}

func Start() error {
	log.Info("Starting Daytona Agent")

	_, err := GetConfig()
	if err != nil {
		return err
	}

	return nil
}
