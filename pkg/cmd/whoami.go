// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/daytonaio/daytona/cmd/daytona/config"

	"github.com/daytonaio/daytona/pkg/cmd/output"
	"github.com/daytonaio/daytona/pkg/views"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const listLabelWidth = 20

var whoamiCmd = &cobra.Command{
	Use:     "whoami",
	Short:   "Display information about the active user",
	Args:    cobra.NoArgs,
	Aliases: []string{"who", "profile active"},
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		profile, err := c.GetActiveProfile()
		if err != nil {
			log.Fatal(err)
		}

		if output.FormatFlag != "" {
			output.Output = profile
			return
		}

		if profile.Id == "default" {
			views.RenderInfoMessageBold("You are currently on the default profile")
		} else {
			views.RenderInfoMessageBold("You are currently on profile " + profile.Name)
		}
		views.RenderListLine(fmt.Sprintf("\x1b[1m%-*s\x1b[0m%s", listLabelWidth, "Profile ID:", profile.Id))

		if profile.Api.Url != "" {
			views.RenderListLine(fmt.Sprintf("\x1b[1m%-*s\x1b[0m%s", listLabelWidth, "API URL:", profile.Api.Url))
		}
	},
}
