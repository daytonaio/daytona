// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	"errors"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/views/profile"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:     "delete [PROFILE_NAME]",
	Short:   "Delete a profile",
	Args:    cobra.RangeArgs(0, 1),
	Aliases: common.GetAliases("delete"),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		var chosenProfileId string
		var chosenProfile *config.Profile

		if len(args) == 0 {
			profilesList := c.Profiles

			chosenProfile, err = profile.GetProfileFromPrompt(profilesList, c.ActiveProfileId, false)
			if err != nil {
				return err
			}

			if chosenProfile == nil {
				return nil
			}

			chosenProfileId = chosenProfile.Id
		} else {
			chosenProfileId = args[0]
		}

		if chosenProfileId == "default" {
			return errors.New("can not delete default profile")
		}

		for _, profile := range c.Profiles {
			if profile.Id == chosenProfileId || profile.Name == chosenProfileId {
				chosenProfile = &profile
				break
			}
		}

		if chosenProfile == nil {
			return errors.New("profile does not exist")
		}

		if c.ActiveProfileId == chosenProfile.Id {
			c.ActiveProfileId = "default"
		}

		for _, profile := range c.Profiles {
			if profile.Name == chosenProfile.Name || profile.Id == chosenProfile.Id {
				err = c.RemoveProfile(profile.Id)
				if err != nil {
					return err
				}
				break
			}
		}

		log.Infof("Deleted profile %s", chosenProfile.Name)
		return nil
	},
}
