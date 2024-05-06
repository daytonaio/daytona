// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	"fmt"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/profile"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var ProfileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage profiles",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

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
	},
}

func init() {
	ProfileCmd.AddCommand(profileListCmd)
	ProfileCmd.AddCommand(ProfileUseCmd)
	ProfileCmd.AddCommand(ProfileAddCmd)
	ProfileCmd.AddCommand(profileEditCmd)
	ProfileCmd.AddCommand(profileDeleteCmd)
}
