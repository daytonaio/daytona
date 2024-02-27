// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	"github.com/daytonaio/daytona/cmd/daytona/config"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var profileUseCmd = &cobra.Command{
	Use:   "use",
	Short: "Use profile [PROFILE_NAME]",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		if len(args) == 0 {
			log.Fatal("Please provide profile name")
		}

		profileArg := args[0]

		var chosenProfile config.Profile

		for _, profile := range c.Profiles {
			if profile.Name == profileArg || profile.Id == profileArg {
				chosenProfile = profile
				break
			}
		}

		if chosenProfile == (config.Profile{}) {
			log.Fatal("Profile does not exist: ", profileArg)
		}

		c.ActiveProfileId = chosenProfile.Id

		err = c.Save()
		if err != nil {
			log.Fatal(err)
		}

		log.Info("Active profile set to ", chosenProfile.Name)
	},
}
