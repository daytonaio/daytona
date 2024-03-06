// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	"fmt"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/views/profile"
	"github.com/daytonaio/daytona/pkg/views/util"

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
			util.RenderInfoMessage("Add a profile by running `daytona profile add")
			return
		}

		if len(profilesList) == 1 {
			util.RenderInfoMessage(fmt.Sprintf("You are using profile %s. Add a new profile by running `daytona profile add`", profilesList[0].Name))
			return
		}

		chosenProfileId := profile.GetProfileIdFromPrompt(profilesList, c.ActiveProfileId, "Choose a profile to use or add a new one", true)

		if chosenProfileId == profile.NewProfileId {
			_, err = CreateProfile(c, nil, true)
			if err != nil {
				log.Fatal(err)
			}
			return
		}

		if chosenProfileId == "" {
			return
		}

		chosenProfile, err := c.GetProfile(chosenProfileId)
		if err != nil {
			log.Fatal(err)
		}

		c.ActiveProfileId = chosenProfile.Id

		err = c.Save()
		if err != nil {
			log.Fatal(err)
		}

		util.RenderInfoMessage(fmt.Sprintf("Active profile set to: %s", chosenProfile.Name))
	},
}

func init() {
	ProfileCmd.AddCommand(profileListCmd)
	ProfileCmd.AddCommand(ProfileUseCmd)
	ProfileCmd.AddCommand(profileAddCmd)
	ProfileCmd.AddCommand(profileEditCmd)
	ProfileCmd.AddCommand(profileDeleteCmd)
}
