// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_profile

import (
	"context"
	"errors"

	"github.com/daytonaio/daytona/cli/api"
	"github.com/daytonaio/daytona/cli/cmd/views/profile/info_view"
	list_view "github.com/daytonaio/daytona/cli/cmd/views/profile/list_view"
	views_provisioner "github.com/daytonaio/daytona/cli/cmd/views/provisioner"
	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/common/api_client"

	view "github.com/daytonaio/daytona/cli/cmd/views/profile/creation_wizard"

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

			chosenProfileId = list_view.GetProfileIdFromPrompt(profilesList, c.ActiveProfileId, "Choose a profile to edit", false)

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

		if profileNameFlag == "" || serverHostnameFlag == "" || serverUserFlag == "" || provisionerFlag == "" || apiUrlFlag == "" || (serverPrivateKeyPathFlag == "" && serverPasswordFlag == "") {
			err = EditProfile(c, true, &chosenProfile)
			if err != nil {
				log.Fatal(err)
			}
			return
		}

		profileEditView := view.ProfileAddView{
			ProfileName:             profileNameFlag,
			RemoteHostname:          serverHostnameFlag,
			RemoteSshPort:           serverPortFlag,
			RemoteSshUser:           serverUserFlag,
			RemoteSshPassword:       serverPasswordFlag,
			RemoteSshPrivateKeyPath: serverPrivateKeyPathFlag,
			DefaultProvisioner:      provisionerFlag,
			ApiUrl:                  apiUrlFlag,
		}

		err = editProfile(chosenProfileId, profileEditView, c, true)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func EditProfile(c *config.Config, notify bool, profile *config.Profile) error {
	if profile == nil {
		return errors.New("profile must not be nil")
	}

	var provisionerPluginList []api_client.ProvisionerPlugin
	var selectedProvisioner *api_client.ProvisionerPlugin = nil
	defaultProvisioner := "default"

	profileAddView := view.ProfileAddView{
		ProfileName:             profile.Name,
		DefaultProvisioner:      defaultProvisioner,
		RemoteSshPassword:       "",
		RemoteSshPrivateKeyPath: "",
		ApiUrl:                  profile.Api.Url,
	}

	if profile.RemoteAuth != nil {
		profileAddView.RemoteSshPort = profile.RemoteAuth.Port
		profileAddView.RemoteHostname = profile.RemoteAuth.Hostname
		profileAddView.RemoteSshUser = profile.RemoteAuth.User

		if profile.RemoteAuth.Password != nil {
			profileAddView.RemoteSshPassword = *profile.RemoteAuth.Password
		} else if profile.RemoteAuth.PrivateKeyPath != nil {
			profileAddView.RemoteSshPrivateKeyPath = *profile.RemoteAuth.PrivateKeyPath
		}
	}

	apiClient, err := api.GetServerApiClient(profile)
	if err != nil {
		log.Fatal(err)
	}

	provisionerPluginList, _, _ = apiClient.PluginAPI.ListProvisionerPlugins(context.Background()).Execute()

	if profile.Id != "default" {
		view.ProfileCreationView(c, &profileAddView, true)
	}

	if len(provisionerPluginList) > 0 {
		for _, provisioner := range provisionerPluginList {
			if *provisioner.Name == profile.DefaultProvisioner {
				selectedProvisioner = &provisioner
				break
			}
		}

		provisioner, err := views_provisioner.GetProvisionerFromPrompt(provisionerPluginList, "Choose a default provisioner to use", selectedProvisioner)
		if err != nil {
			return err
		}

		if profile.Id == "default" {
			profile.DefaultProvisioner = *provisioner.Name
			return c.EditProfile(*profile)
		} else {
			profileAddView.DefaultProvisioner = *provisionerPluginList[0].Name
		}
	}

	return editProfile(profile.Id, profileAddView, c, notify)
}

func editProfile(profileId string, profileView view.ProfileAddView, c *config.Config, notify bool) error {
	editedProfile := config.Profile{
		Id:   profileId,
		Name: profileView.ProfileName,
		RemoteAuth: &config.RemoteAuth{
			Port:           profileView.RemoteSshPort,
			Hostname:       profileView.RemoteHostname,
			User:           profileView.RemoteSshUser,
			Password:       nil,
			PrivateKeyPath: nil,
		},
		DefaultProvisioner: profileView.DefaultProvisioner,
		Api: config.ServerApi{
			Url: profileView.ApiUrl,
		},
	}
	if profileView.RemoteSshPassword != "" {
		editedProfile.RemoteAuth.Password = &profileView.RemoteSshPassword
	} else if profileView.RemoteSshPrivateKeyPath != "" {
		editedProfile.RemoteAuth.PrivateKeyPath = &profileView.RemoteSshPrivateKeyPath
	} else {
		return errors.New("password or private key path must be provided")
	}

	err := c.EditProfile(editedProfile)
	if err != nil {
		return err
	}

	if notify {
		info_view.Render(info_view.ProfileInfo{
			ProfileName: profileView.ProfileName,
			ApiUrl:      profileView.RemoteHostname,
		}, "edited")
	}

	return nil
}

func init() {
	profileEditCmd.Flags().StringVarP(&profileNameFlag, "name", "n", "", "Profile name")
	profileEditCmd.Flags().StringVarP(&serverHostnameFlag, "hostname", "h", "", "Remote SSH hostname")
	profileEditCmd.Flags().IntVarP(&serverPortFlag, "port", "P", 22, "Remote SSH port")
	profileEditCmd.Flags().StringVarP(&serverUserFlag, "user", "u", "", "Remote SSH url")
	profileEditCmd.Flags().StringVarP(&serverPasswordFlag, "password", "p", "", "Remote SSH password")
	profileEditCmd.Flags().StringVarP(&serverPrivateKeyPathFlag, "private-key-path", "k", "", "Remote SSH private key path")
	profileEditCmd.Flags().StringVarP(&provisionerFlag, "provisioner", "r", "default", "Provisioner")
	profileEditCmd.Flags().StringVarP(&apiUrlFlag, "api-url", "a", "", "API URL")
}
