// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	view "github.com/daytonaio/daytona/pkg/views/server"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/daytonaio/daytona/pkg/cmd/format"
	"github.com/daytonaio/daytona/pkg/server"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Output local Daytona Server config",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := server.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		if format.FormatFlag != "" {
			formatter := format.NewFormatter(config)
			formatter.Print()
			return
		}

		view.RenderConfig(config)
	},
}

func init() {
	format.RegisterFormatFlag(configCmd)
}
