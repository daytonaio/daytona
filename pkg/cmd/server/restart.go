// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/daytonaio/daytona/pkg/cmd/server/daemon"
	"github.com/daytonaio/daytona/pkg/views"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restarts the Daytona Server daemon",
	Run: func(cmd *cobra.Command, args []string) {
		views.RenderInfoMessage("Stopping the Daytona Server daemon...")
		err := daemon.Stop()
		if err != nil {
			log.Fatal(err)
		}

		views.RenderInfoMessage("Starting the Daytona Server daemon...")
		err = daemon.Start()
		if err != nil {
			log.Fatal(err)
		}
		views.RenderInfoMessageBold("Daytona Server daemon restarted successfully")
	},
}
