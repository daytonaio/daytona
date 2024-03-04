// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	"context"
	"errors"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views/profile"
	"github.com/daytonaio/daytona/pkg/views/provider"
	"github.com/daytonaio/daytona/pkg/views/provider/target"

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

		if profileNameFlag == "" || serverHostnameFlag == "" || serverUserFlag == "" || providerFlag == "" || providerTargetFlag == "" || apiUrlFlag == "" || (serverPrivateKeyPathFlag == "" && serverPasswordFlag == "") {
			err = EditProfile(c, true, &chosenProfile, false, false)
			if err != nil {
				log.Fatal(err)
			}
			return
		}

		profileEditView := profile.ProfileAddView{
			ProfileName:             profileNameFlag,
			RemoteHostname:          serverHostnameFlag,
			RemoteSshPort:           serverPortFlag,
			RemoteSshUser:           serverUserFlag,
			RemoteSshPassword:       serverPasswordFlag,
			RemoteSshPrivateKeyPath: serverPrivateKeyPathFlag,
			DefaultProviderName:     providerFlag,
			DefaultProviderTarget:   providerTargetFlag,
			ApiUrl:                  apiUrlFlag,
		}

		err = editProfile(&chosenProfile, profileEditView, c, true)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func EditProfile(c *config.Config, notify bool, profileToEdit *config.Profile, forceRemoteAccess, skipName bool) error {
	if profileToEdit == nil {
		return errors.New("profile must not be nil")
	}

	var providersList []serverapiclient.Provider

	profileAddView := profile.ProfileAddView{
		ProfileName:             profileToEdit.Name,
		DefaultProviderName:     "",
		DefaultProviderTarget:   "",
		RemoteSshPort:           22,
		RemoteSshPassword:       "",
		RemoteSshPrivateKeyPath: "",
		ApiUrl:                  profileToEdit.Api.Url,
	}

	if profileToEdit.RemoteAuth != nil {
		profileAddView.RemoteSshPort = profileToEdit.RemoteAuth.Port
		profileAddView.RemoteHostname = profileToEdit.RemoteAuth.Hostname
		profileAddView.RemoteSshUser = profileToEdit.RemoteAuth.User

		if profileToEdit.RemoteAuth.Password != nil {
			profileAddView.RemoteSshPassword = *profileToEdit.RemoteAuth.Password
		} else if profileToEdit.RemoteAuth.PrivateKeyPath != nil {
			profileAddView.RemoteSshPrivateKeyPath = *profileToEdit.RemoteAuth.PrivateKeyPath
		}
	}

	apiClient, err := server.GetApiClient(profileToEdit)
	if err != nil {
		log.Fatal(err)
	}

	providersList, _, _ = apiClient.ProviderAPI.ListProviders(context.Background()).Execute()

	if profileToEdit.Id != "default" {
		profile.ProfileCreationView(c, &profileAddView, true, forceRemoteAccess, skipName)
	}

	if len(providersList) > 0 {
		provider := provider.GetProviderFromPrompt(providersList, "Choose a default provider to use")

		defaultTarget, err := target.GetTargetFromPrompt(provider.Targets, false)
		if err != nil {
			return err
		}

		if profileToEdit.Id == "default" {
			profileToEdit.DefaultProvider = &config.DefaultProvider{
				Provider: *provider.Name,
				Target:   *defaultTarget.Name,
			}
			return c.EditProfile(*profileToEdit)
		} else {
			profileAddView.DefaultProviderName = *providersList[0].Name
			profileAddView.DefaultProviderTarget = *defaultTarget.Name
		}
	}

	return editProfile(profileToEdit, profileAddView, c, notify)
}

func editProfile(profileToEdit *config.Profile, profileView profile.ProfileAddView, c *config.Config, notify bool) error {
	profileToEdit.Name = profileView.ProfileName
	profileToEdit.Api.Url = profileView.ApiUrl
	profileToEdit.DefaultProvider = &config.DefaultProvider{
		Provider: profileView.DefaultProviderName,
		Target:   profileView.DefaultProviderTarget,
	}
	profileToEdit.RemoteAuth = &config.RemoteAuth{
		Port:     profileView.RemoteSshPort,
		Hostname: profileView.RemoteHostname,
		User:     profileView.RemoteSshUser,
	}
	profileToEdit.Api = config.ServerApi{
		Url: profileView.ApiUrl,
	}

	if profileView.RemoteSshPassword != "" {
		profileToEdit.RemoteAuth.Password = &profileView.RemoteSshPassword
	} else if profileView.RemoteSshPrivateKeyPath != "" {
		profileToEdit.RemoteAuth.PrivateKeyPath = &profileView.RemoteSshPrivateKeyPath
	} else {
		return errors.New("password or private key path must be provided")
	}

	err := c.EditProfile(*profileToEdit)
	if err != nil {
		return err
	}

	if notify {
		profile.Render(profile.ProfileInfo{
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
	profileEditCmd.Flags().StringVarP(&providerFlag, "provider", "r", "default", "Provider")
	profileEditCmd.Flags().StringVarP(&apiUrlFlag, "api-url", "a", "", "API URL")
}
