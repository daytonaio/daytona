// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/views/profile"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var profileDeleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete profile [PROFILE_NAME]",
	Args:    cobra.RangeArgs(0, 1),
	Aliases: []string{"remove", "rm"},
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		var chosenProfile *config.Profile

		if len(args) == 0 {
			profilesList := c.Profiles

			chosenProfile, err = profile.GetProfileFromPrompt(profilesList, c.ActiveProfileId, false)
			if err != nil {
				log.Fatal(err)
			}

			if chosenProfile.Id == "" {
				return
			}
		} else {
			chosenProfile.Id = args[0]
		}

		if chosenProfile.Id == "default" {
			log.Fatal("Can not delete default profile")
		}

		for _, profile := range c.Profiles {
			if profile.Id == chosenProfile.Id || profile.Name == chosenProfile.Id {
				chosenProfile = &profile
				break
			}
		}

		if chosenProfile == nil {
			log.Fatal("Profile does not exist")
			return
		}

		if c.ActiveProfileId == chosenProfile.Id {
			c.ActiveProfileId = "default"
		}

		for _, profile := range c.Profiles {
			if profile.Name == chosenProfile.Name || profile.Id == chosenProfile.Id {
				err = c.RemoveProfile(profile.Id)
				if err != nil {
					log.Fatal(err)
				}
				break
			}
		}

		log.Infof("Deleted profile %s", chosenProfile.Name)
	},
}
