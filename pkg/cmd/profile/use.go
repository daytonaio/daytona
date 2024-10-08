// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	"fmt"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"

	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/profile"
	"github.com/spf13/cobra"
)

var ProfileUseCmd = &cobra.Command{
	Use:     "use",
	Short:   "Use profile [PROFILE_NAME]",
	Args:    cobra.MaximumNArgs(1),
	GroupID: util.PROFILE_GROUP,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		if len(args) == 0 {
			profilesList := c.Profiles

			if len(profilesList) == 0 {
				views.RenderInfoMessage("Add a profile by running `daytona profile add`")
				return nil
			}

			if len(profilesList) == 1 {
				views.RenderInfoMessage(fmt.Sprintf("You are using profile %s. Add a new profile by running `daytona profile add`", profilesList[0].Name))
				return nil
			}

			chosenProfile, err := profile.GetProfileFromPrompt(profilesList, c.ActiveProfileId, true)
			if err != nil {
				return err
			}

			if chosenProfile == nil {
				return nil
			}

			if chosenProfile.Id == profile.NewProfileId {
				_, err = CreateProfile(c, nil, true)
				return err
			}

			if chosenProfile.Id == "" {
				return nil
			}

			profile, err := c.GetProfile(chosenProfile.Id)
			if err != nil {
				return err
			}

			c.ActiveProfileId = profile.Id

			err = c.Save()
			if err != nil {
				return err
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
				return fmt.Errorf("profile does not exist: %s", profileArg)
			}

			c.ActiveProfileId = chosenProfile.Id

			err = c.Save()
			if err != nil {
				return err
			}

			views.RenderInfoMessage(fmt.Sprintf("Active profile set to: %s", chosenProfile.Name))
		}
		return nil
	},
}
