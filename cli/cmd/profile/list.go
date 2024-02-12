// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_profile

import (
	"github.com/daytonaio/daytona/cli/cmd/output"
	list_view "github.com/daytonaio/daytona/cli/cmd/views/profile/list_view"
	"github.com/daytonaio/daytona/cli/config"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var profileListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List profiles",
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		if output.FormatFlag != "" {
			output.Output = c.Profiles
			return
		}

		list_view.ListProfiles(c.Profiles, c.ActiveProfileId)
	},
}
