//go:build !windows

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"github.com/daytonaio/daytona/pkg/agent"
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

		err := agent.Start()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {

}
