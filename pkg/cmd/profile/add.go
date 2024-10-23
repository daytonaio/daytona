// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/views/profile"

	"github.com/spf13/cobra"
)

var ProfileAddCmd = &cobra.Command{
	Use:     "add",
	Short:   "Add profile",
	Args:    cobra.NoArgs,
	Aliases: []string{"new"},
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		profileAddView := profile.ProfileAddView{
			ProfileName: profileNameFlag,
			ApiUrl:      apiUrlFlag,
			ApiKey:      apiKeyFlag,
		}

		if profileAddView.ProfileName != "" && profileAddView.ApiUrl != "" && profileAddView.ApiKey != "" {
			_, err = addProfile(profileAddView, c, true)
		} else {
			_, err = CreateProfile(c, &profileAddView, true)
		}

		return err
	},
}

func CreateProfile(c *config.Config, profileAddView *profile.ProfileAddView, notify bool) (string, error) {
	if profileAddView == nil {
		profileAddView = &profile.ProfileAddView{
			ProfileName: "",
			ApiUrl:      "",
			ApiKey:      "",
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
			Key: profileView.ApiKey,
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
var apiKeyFlag string

func init() {
	ProfileAddCmd.Flags().StringVarP(&profileNameFlag, "name", "n", "", "Profile name")
	ProfileAddCmd.Flags().StringVarP(&apiUrlFlag, "api-url", "a", "", "API URL")
	ProfileAddCmd.Flags().StringVarP(&apiKeyFlag, "api-key", "k", "", "API Key")
}
