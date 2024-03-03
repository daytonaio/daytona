// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	"fmt"

	"github.com/daytonaio/daytona/cmd/daytona/config"

	log "github.com/sirupsen/logrus"

	"github.com/daytonaio/daytona/pkg/cmd/output"
	view_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/spf13/cobra"
)

var ProfileUseCmd = &cobra.Command{
	Use:   "use",
	Short: "Set the active profile",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		if len(args) == 0 {
			log.Fatal("Please provide the profile name")
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

		if output.FormatFlag != "" {
			output.Output = chosenProfile.Id
			return
		}

		view_util.RenderInfoMessage(fmt.Sprintf("Active profile set to %s", chosenProfile.Name))
	},
}
