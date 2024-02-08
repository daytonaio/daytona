// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_server

import (
	"github.com/daytonaio/daytona/server"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var startupCmd = &cobra.Command{
	Use:   "startup",
	Short: "Runs Daytona Server in the background",
	Run: func(cmd *cobra.Command, args []string) {
		err := server.StartDaemon()
		if err != nil {
			log.Fatal(err)
		}
	},
}
