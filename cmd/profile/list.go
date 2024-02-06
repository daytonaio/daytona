// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_profile

import (
	profile_list "github.com/daytonaio/daytona/cmd/views/profilie_list"
	"github.com/daytonaio/daytona/config"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List profiles",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		profile_list.Render(c.Profiles, c.ActiveProfileId)
	},
}
