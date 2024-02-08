// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_profile

import (
	"fmt"

	profile_list "github.com/daytonaio/daytona/cmd/views/profile_list"
	"github.com/daytonaio/daytona/config"
	"github.com/daytonaio/daytona/output"

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

		chosenProfileId := profile_list.GetProfileIdFromPrompt(c.Profiles, c.ActiveProfileId, "Profiles", false)

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

		output.Output = c.Profiles
	},
}
