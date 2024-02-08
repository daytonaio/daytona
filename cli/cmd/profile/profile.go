// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_profile

import (
	"fmt"

	profile_list "github.com/daytonaio/daytona/cli/cmd/views/profile_list"
	views_util "github.com/daytonaio/daytona/cli/cmd/views/util"
	"github.com/daytonaio/daytona/cli/config"

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
			views_util.RenderInfoMessage("Add a profile by running `daytona profile add")
			return
		}

		if len(profilesList) == 1 {
			views_util.RenderInfoMessage(fmt.Sprintf("You are using profile %s. Add a new profile by running `daytona profile add`", profilesList[0].Name))
			return
		}

		chosenProfileId := profile_list.GetProfileIdFromPrompt(profilesList, c.ActiveProfileId, "Choose a profile to use or add a new one", true)

		if chosenProfileId == profile_list.NewProfileId {
			CreateProfile(c, true)
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

		views_util.RenderInfoMessage(fmt.Sprintf("Active profile set to: %s", chosenProfile.Name))
	},
}

func init() {
	ProfileCmd.AddCommand(profileListCmd)
	ProfileCmd.AddCommand(profileUseCmd)
	ProfileCmd.AddCommand(profileAddCmd)
	ProfileCmd.AddCommand(profileEditCmd)
	ProfileCmd.AddCommand(profileDeleteCmd)
}
