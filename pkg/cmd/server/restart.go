// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/daytonaio/daytona/pkg/cmd/server/daemon"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restarts the Daytona Server daemon",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Stopping the Daytona Server daemon...")
		err := daemon.Stop()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Starting the Daytona Server daemon...")
		err = daemon.Start()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Daytona Server daemon restarted successfully")
	},
}
