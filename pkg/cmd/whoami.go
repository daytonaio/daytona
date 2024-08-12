// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"

	"github.com/daytonaio/daytona/pkg/cmd/output"
	view "github.com/daytonaio/daytona/pkg/views/whoami"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var formatFlag string
var whoamiCmd = &cobra.Command{
	Use:     "whoami",
	Short:   "Display information about the active user",
	Args:    cobra.NoArgs,
	Aliases: []string{"who", "user"},
	GroupID: util.PROFILE_GROUP,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		profile, err := c.GetActiveProfile()
		if err != nil {
			log.Fatal(err)
		}

		if formatFlag != "" {
			display := output.NewOutputFormatter(profile, formatFlag)
			display.Print()
			return
		}

		view.Render(profile)
	},
}

func init() {
	whoamiCmd.PersistentFlags().StringVarP(&formatFlag, output.FormatFlagName, output.FormatFlagShortHand, formatFlag, output.FormatDescription)
	whoamiCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if formatFlag != "" {
			output.BlockStdOut()
		}
	}
}
