// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	config_view "github.com/daytonaio/daytona/pkg/views/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var showApiKeysFlag bool

var configCmd = &cobra.Command{
	Use:     "config",
	Short:   "Output Daytona configuration",
	Aliases: []string{"cfg"},
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
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
			return
		}

		config_view.Render(c, showApiKeysFlag)
	},
}

func init() {
	format.RegisterFormatFlag(configCmd)
	configCmd.Flags().BoolVarP(&showApiKeysFlag, "show-api-keys", "k", false, "Show API keys")
}
