// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/views/profile"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var profileAddCmd = &cobra.Command{
	Use:     "add",
	Short:   "Add profile",
	Args:    cobra.NoArgs,
	Aliases: []string{"new"},
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		profileAddView := profile.ProfileAddView{
			ProfileName: profileNameFlag,
			ApiUrl:      apiUrlFlag,
		}

		if profileAddView.ProfileName != "" && profileAddView.ApiUrl != "" {
			_, err = addProfile(profileAddView, c, true)
		}

		if profileNameFlag == "" || apiUrlFlag == "" {
			_, err = CreateProfile(c, nil, true)
		}

		if err != nil {
			log.Fatal(err)
		}
	},
}

func CreateProfile(c *config.Config, profileAddView *profile.ProfileAddView, notify bool) (string, error) {
	if profileAddView == nil {
		profileAddView = &profile.ProfileAddView{
			ProfileName: "",
			ApiUrl:      "",
		}
	}

	profile.ProfileCreationView(c, profileAddView, false)

	return addProfile(*profileAddView, c, notify)
}

func addProfile(profileView profile.ProfileAddView, c *config.Config, notify bool) (string, error) {
	newProfile := config.Profile{
		Id:   util.GenerateIdFromName(profileView.ProfileName),
		Name: profileView.ProfileName,
		Api: config.ServerApi{
			Url: profileView.ApiUrl,
		},
	}

	newProfile.Api.Url = profileView.ApiUrl
	err := c.AddProfile(newProfile)
	if err != nil {
		return "", err
	}

	if notify {
		profile.Render(profile.ProfileInfo{
			ProfileName: newProfile.Name,
			ApiUrl:      newProfile.Api.Url,
		}, "added and set as active")
	}

	return newProfile.Id, nil
}

var profileNameFlag string
var apiUrlFlag string

func init() {
	profileAddCmd.Flags().StringVarP(&profileNameFlag, "name", "n", "", "Profile name")
	profileAddCmd.Flags().StringVarP(&apiUrlFlag, "api-url", "a", "", "API URL")
}
