// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	view "github.com/daytonaio/daytona/pkg/views/server"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/daytonaio/daytona/pkg/cmd/output"
	"github.com/daytonaio/daytona/pkg/server"
)

var formatFlag string
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Output local Daytona Server config",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := server.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		if formatFlag != "" {
			display := output.NewOutputFormatter(config, formatFlag)
			display.Print()
			return
		}

		view.RenderConfig(config)
	},
}

func init() {
	configCmd.PersistentFlags().StringVarP(&formatFlag, output.FormatFlagName, output.FormatFlagShortHand, formatFlag, output.FormatDescription)
	configCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if formatFlag != "" {
			output.BlockStdOut()
		}
	}
}
