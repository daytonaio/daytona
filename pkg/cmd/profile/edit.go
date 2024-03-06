// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	"errors"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/views/profile"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var profileEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit profile [PROFILE_NAME]",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		var chosenProfileId string
		var chosenProfile config.Profile

		if len(args) == 0 {
			profilesList := c.Profiles

			chosenProfileId = profile.GetProfileIdFromPrompt(profilesList, c.ActiveProfileId, "Choose a profile to edit", false)

			if chosenProfileId == "" {
				return
			}
		} else {
			chosenProfileId = args[0]
		}

		for _, profile := range c.Profiles {
			if profile.Id == chosenProfileId || profile.Name == chosenProfileId {
				chosenProfile = profile
				break
			}
		}

		if chosenProfile == (config.Profile{}) {
			log.Fatal("Profile does not exist")
		}

		if profileNameFlag == "" || apiUrlFlag == "" {
			err = EditProfile(c, true, &chosenProfile)
			if err != nil {
				log.Fatal(err)
			}
			return
		}

		profileEditView := profile.ProfileAddView{
			ProfileName: profileNameFlag,
			ApiUrl:      apiUrlFlag,
		}

		err = editProfile(&chosenProfile, profileEditView, c, true)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func EditProfile(c *config.Config, notify bool, profileToEdit *config.Profile) error {
	if profileToEdit == nil {
		return errors.New("profile must not be nil")
	}

	profileAddView := profile.ProfileAddView{
		ProfileName: profileToEdit.Name,
		ApiUrl:      profileToEdit.Api.Url,
	}

	if profileToEdit.Id != "default" {
		profile.ProfileCreationView(c, &profileAddView, true)
	}

	return editProfile(profileToEdit, profileAddView, c, notify)
}

func editProfile(profileToEdit *config.Profile, profileView profile.ProfileAddView, c *config.Config, notify bool) error {
	profileToEdit.Name = profileView.ProfileName
	profileToEdit.Api = config.ServerApi{
		Url: profileView.ApiUrl,
	}

	err := c.EditProfile(*profileToEdit)
	if err != nil {
		return err
	}

	if notify {
		profile.Render(profile.ProfileInfo{
			ProfileName: profileView.ProfileName,
			ApiUrl:      profileView.ApiUrl,
		}, "edited")
	}

	return nil
}

func init() {
	profileEditCmd.Flags().StringVarP(&profileNameFlag, "name", "n", "", "Profile name")
	profileEditCmd.Flags().StringVarP(&apiUrlFlag, "api-url", "a", "", "API URL")
}
