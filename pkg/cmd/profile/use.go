// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	"fmt"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"

	log "github.com/sirupsen/logrus"

	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/profile"
	"github.com/spf13/cobra"
)

var ProfileUseCmd = &cobra.Command{
	Use:     "use",
	Short:   "Use profile [PROFILE_NAME]",
	Args:    cobra.MaximumNArgs(1),
	GroupID: util.PROFILE_GROUP,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		if len(args) == 0 {
			profilesList := c.Profiles

			if len(profilesList) == 0 {
				views.RenderInfoMessage("Add a profile by running `daytona profile add`")
				return
			}

			if len(profilesList) == 1 {
				views.RenderInfoMessage(fmt.Sprintf("You are using profile %s. Add a new profile by running `daytona profile add`", profilesList[0].Name))
				return
			}

			chosenProfile, err := profile.GetProfileFromPrompt(profilesList, c.ActiveProfileId, true)
			if err != nil {
				log.Fatal(err)
			}

			if chosenProfile == nil {
				return
			}

			if chosenProfile.Id == profile.NewProfileId {
				_, err = CreateProfile(c, nil, true)
				if err != nil {
					log.Fatal(err)
				}
				return
			}

			if chosenProfile.Id == "" {
				return
			}

			profile, err := c.GetProfile(chosenProfile.Id)
			if err != nil {
				log.Fatal(err)
			}

			c.ActiveProfileId = profile.Id

			err = c.Save()
			if err != nil {
				log.Fatal(err)
			}

			views.RenderInfoMessage(fmt.Sprintf("Active profile set to: %s", profile.Name))
		} else {
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

			views.RenderInfoMessage(fmt.Sprintf("Active profile set to: %s", chosenProfile.Name))
		}
	},
}
