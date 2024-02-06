// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_agent

import (
	agent_config "dagent/agent/config"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure Daytona Agent",
	Run: func(cmd *cobra.Command, args []string) {
		err := agent_config.Configure()
		if err != nil {
			log.Fatal(err)
		}
	},
}
