// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	view "github.com/daytonaio/daytona/pkg/views/server"
	"github.com/spf13/cobra"

	"github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	"github.com/daytonaio/daytona/pkg/server"
)

var configCmd = &cobra.Command{
	Use:     "config",
	Short:   "Output local Daytona Server config",
	Aliases: common.GetAliases("config"),
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := server.GetConfig()
		if err != nil {
			return err
		}

		if format.FormatFlag != "" {
			formattedData := format.NewFormatter(config)
			formattedData.Print()
			return nil
		}

		view.RenderConfig(config)
		return nil
	},
}

func init() {
	format.RegisterFormatFlag(configCmd)
}
