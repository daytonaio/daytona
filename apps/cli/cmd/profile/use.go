// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package profile

import (
	"fmt"

	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/cli/views/common"
	"github.com/daytonaio/daytona/cli/views/profile"
	"github.com/spf13/cobra"
)

var UseCmd = &cobra.Command{
	Use:   "use [PROFILE]",
	Short: "Set active profile",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		if len(c.Profiles) == 0 {
			return config.ErrNoProfilesFound
		}

		var chosenProfile config.Profile

		if len(args) == 0 {
			chosenProfilePtr, err := profile.GetProfileFromPrompt(c.Profiles)
			if err != nil {
				return err
			}
			if chosenProfilePtr == nil {
				return nil
			}
			chosenProfile = *chosenProfilePtr
		} else {
			profileIdOrName := args[0]
			found := false
			for _, p := range c.Profiles {
				if p.Id == profileIdOrName || p.Name == profileIdOrName {
					chosenProfile = p
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("profile %s not found", profileIdOrName)
			}
		}

		err = c.SetActiveProfile(chosenProfile.Id)
		if err != nil {
			return err
		}

		common.RenderInfoMessageBold(fmt.Sprintf("Profile %s is now active", chosenProfile.Name))
		return nil
	},
}
