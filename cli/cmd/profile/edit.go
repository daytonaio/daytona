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
	"github.com/daytonaio/daytona/cli/connection"
	"github.com/daytonaio/daytona/common/api_client"
	"google.golang.org/grpc"

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

		if profileNameFlag == "" || serverHostnameFlag == "" || serverUserFlag == "" || provisionerFlag == "" || (serverPrivateKeyPathFlag == "" && serverPasswordFlag == "") {
			conn, err := connection.GetGrpcConn(nil)
			if err != nil {
				log.Fatal(err)
			}

			err = EditProfile(c, conn, true, &chosenProfile)
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
		}

		err = editProfile(chosenProfileId, profileEditView, c, true)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func EditProfile(c *config.Config, conn *grpc.ClientConn, notify bool, profile *config.Profile) error {
	if profile == nil {
		return errors.New("profile must not be nil")
	}

	apiClient := api.GetServerApiClient("http://localhost:3000", "")

	provisionerPluginList, _, err := apiClient.PluginAPI.ListProvisionerPlugins(context.Background()).Execute()
	if err != nil {
		log.Fatal(err)
	}

	if len(provisionerPluginList) == 0 {
		return errors.New("no provisioner plugins found")
	}

	var selectedProvisioner *api_client.ProvisionerPlugin = nil
	for _, provisioner := range provisionerPluginList {
		if *provisioner.Name == profile.DefaultProvisioner {
			selectedProvisioner = &provisioner
			break
		}
	}

	if profile.Id == "default" {
		provisioner, err := views_provisioner.GetProvisionerFromPrompt(provisionerPluginList, "Choose a default provisioner to use", selectedProvisioner)
		if err != nil {
			return err
		}
		profile.DefaultProvisioner = *provisioner.Name
		return c.EditProfile(*profile)
	}

	profileAddView := view.ProfileAddView{
		ProfileName:             profile.Name,
		RemoteHostname:          profile.Hostname,
		RemoteSshPort:           profile.Port,
		RemoteSshUser:           profile.Auth.User,
		DefaultProvisioner:      *provisionerPluginList[0].Name,
		RemoteSshPassword:       "",
		RemoteSshPrivateKeyPath: "",
	}

	if profile.Auth.Password != nil {
		profileAddView.RemoteSshPassword = *profile.Auth.Password
	} else if profile.Auth.PrivateKeyPath != nil {
		profileAddView.RemoteSshPrivateKeyPath = *profile.Auth.PrivateKeyPath
	}

	view.ProfileCreationView(c, &profileAddView, true)

	provisioner, err := views_provisioner.GetProvisionerFromPrompt(provisionerPluginList, "Choose a provisioner to use", selectedProvisioner)
	if err != nil {
		return err
	}
	profileAddView.DefaultProvisioner = *provisioner.Name

	return editProfile(profile.Id, profileAddView, c, notify)
}

func editProfile(profileId string, profileView view.ProfileAddView, c *config.Config, notify bool) error {
	editedProfile := config.Profile{
		Id:       profileId,
		Name:     profileView.ProfileName,
		Hostname: profileView.RemoteHostname,
		Port:     profileView.RemoteSshPort,
		Auth: config.ProfileAuth{
			User:           profileView.RemoteSshUser,
			Password:       nil,
			PrivateKeyPath: nil,
		},
		DefaultProvisioner: profileView.DefaultProvisioner,
	}
	if profileView.RemoteSshPassword != "" {
		editedProfile.Auth.Password = &profileView.RemoteSshPassword
	} else if profileView.RemoteSshPrivateKeyPath != "" {
		editedProfile.Auth.PrivateKeyPath = &profileView.RemoteSshPrivateKeyPath
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
			ServerUrl:   profileView.RemoteHostname,
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
}
