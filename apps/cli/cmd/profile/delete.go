// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package profile

import (
	"fmt"

	"github.com/daytonaio/daytona/cli/config"
	view_common "github.com/daytonaio/daytona/cli/views/common"
	"github.com/daytonaio/daytona/cli/views/profile"
	"github.com/spf13/cobra"
)

var DeleteCmd = &cobra.Command{
	Use:   "delete [PROFILE]",
	Short: "Delete a profile",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		if len(c.Profiles) == 0 {
			return config.ErrNoProfilesFound
		}

		var profileToDelete config.Profile

		if len(args) == 0 {
			profilePtr, err := profile.GetProfileFromPrompt(c.Profiles)
			if err != nil {
				return err
			}
			if profilePtr == nil {
				return nil
			}
			profileToDelete = *profilePtr
		} else {
			profileIdOrName := args[0]
			found := false
			for _, p := range c.Profiles {
				if p.Id == profileIdOrName || p.Name == profileIdOrName {
					profileToDelete = p
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("profile %s not found", profileIdOrName)
			}
		}

		err = c.RemoveProfile(profileToDelete.Id)
		if err != nil {
			return err
		}

		view_common.RenderInfoMessageBold(fmt.Sprintf("Profile %s deleted", profileToDelete.Name))
		return nil
	},
}
