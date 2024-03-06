// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/server/config"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the server process",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if log.GetLevel() < log.InfoLevel {
			//	for now, force the log level to info when running the server
			log.SetLevel(log.InfoLevel)
		}

		config, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		if config == nil {
			log.Fatal("Server configuration is not set. Please run `daytona configure`.")
		}

		err = server.Start()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	ServerCmd.AddCommand(configureCmd)
	ServerCmd.AddCommand(configCmd)
	ServerCmd.AddCommand(logsCmd)
}
