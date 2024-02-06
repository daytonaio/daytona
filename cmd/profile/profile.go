// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_profile

import (
	profile_list "dagent/cmd/views/profilie_list"
	"dagent/config"
	"fmt"

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
			fmt.Println("Add a profile by running `daytona profile add`")
			return
		}

		if len(profilesList) == 1 {
			fmt.Println("You are using profile " + profilesList[0].Name + ". Add a new profile by running `daytona profile add`")
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

		fmt.Printf("\nActive profile set to: %s\n\n", chosenProfile.Name)
	},
}

func init() {
	ProfileCmd.AddCommand(profileListCmd)
	ProfileCmd.AddCommand(profileUseCmd)
	ProfileCmd.AddCommand(profileAddCmd)
	ProfileCmd.AddCommand(profileEditCmd)
	ProfileCmd.AddCommand(profileDeleteCmd)
}
