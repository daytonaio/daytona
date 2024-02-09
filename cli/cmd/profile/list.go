// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_profile

import (
	"fmt"

	"github.com/daytonaio/daytona/cli/cmd/output"
	list_view "github.com/daytonaio/daytona/cli/cmd/views/profile/list_view"
	views_util "github.com/daytonaio/daytona/cli/cmd/views/util"
	"github.com/daytonaio/daytona/cli/config"

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

		if output.FormatFlag != "" {
			output.Output = c.Profiles
			return
		}

		chosenProfileId := list_view.GetProfileIdFromPrompt(c.Profiles, c.ActiveProfileId, "Profiles", false)

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
