// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_agent

import (
	"github.com/daytonaio/daytona/agent"
	"github.com/daytonaio/daytona/agent/config"
	cmd_agent_key "github.com/daytonaio/daytona/cmd/agent/key"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var AgentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Start the agent process",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if log.GetLevel() < log.InfoLevel {
			//	for now, force the log level to info when running the agent
			log.SetLevel(log.InfoLevel)
		}

		config, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		if config == nil {
			log.Fatal("Agent configuration is not set. Please run `daytona configure`.")
		}

		err = agent.Start()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	AgentCmd.AddCommand(cmd_agent_key.KeyCmd)
	AgentCmd.AddCommand(configureCmd)
	AgentCmd.AddCommand(startupCmd)
	AgentCmd.AddCommand(installCmd)
	AgentCmd.AddCommand(uninstallCmd)
}
