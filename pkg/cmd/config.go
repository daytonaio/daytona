// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	config_view "github.com/daytonaio/daytona/pkg/views/config"
	"github.com/spf13/cobra"
)

var showApiKeysFlag bool

var configCmd = &cobra.Command{
	Use:     "config",
	Short:   "Output Daytona configuration",
	Aliases: []string{"cfg"},
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		if format.FormatFlag != "" {
			// Hide API keys if not explicitly requested
			if !showApiKeysFlag {
				for i := range c.Profiles {
					c.Profiles[i].Api.Key = "*********************"
				}
			}

			formattedData := format.NewFormatter(&c)
			formattedData.Print()
			return nil
		}

		config_view.Render(c, showApiKeysFlag)
		return nil
	},
}

func init() {
	format.RegisterFormatFlag(configCmd)
	configCmd.Flags().BoolVarP(&showApiKeysFlag, "show-api-keys", "k", false, "Show API keys")
}
