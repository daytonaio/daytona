// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_profile

import (
	"errors"

	"github.com/daytonaio/daytona/cli/cmd/views/profile/info_view"
	list_view "github.com/daytonaio/daytona/cli/cmd/views/profile/list_view"
	"github.com/daytonaio/daytona/cli/config"

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

		if chosenProfileId == "default" {
			log.Fatal("Can not edit default profile")
		}

		for _, profile := range c.Profiles {
			if profile.Id == chosenProfileId || profile.Name == chosenProfileId {
				chosenProfile = profile
				break
			}
		}

		if chosenProfile == (config.Profile{}) {
			log.Fatal("Profile does not exist")
			return
		}

		profileEditView := view.ProfileAddView{
			ProfileName:    chosenProfile.Name,
			RemoteHostname: chosenProfile.Hostname,
			RemoteSshPort:  chosenProfile.Port,
			RemoteSshUser:  chosenProfile.Auth.User,
		}

		if chosenProfile.Auth.Password != nil {
			profileEditView.RemoteSshPassword = *chosenProfile.Auth.Password
			profileEditView.RemoteSshPrivateKeyPath = ""
		} else if chosenProfile.Auth.PrivateKeyPath != nil {
			profileEditView.RemoteSshPassword = ""
			profileEditView.RemoteSshPrivateKeyPath = *chosenProfile.Auth.PrivateKeyPath
		}

		if profileNameFlag == "" || serverHostnameFlag == "" || serverUserFlag == "" || (serverPrivateKeyPathFlag == "" && serverPasswordFlag == "") {
			view.ProfileCreationView(c, &profileEditView, true)
		} else {
			profileEditView.ProfileName = profileNameFlag
			profileEditView.RemoteHostname = serverHostnameFlag
			profileEditView.RemoteSshPassword = serverPasswordFlag
			profileEditView.RemoteSshUser = serverUserFlag
			profileEditView.RemoteSshPrivateKeyPath = serverPrivateKeyPathFlag
		}

		editedProfile := config.Profile{
			Id:       chosenProfileId,
			Name:     profileEditView.ProfileName,
			Hostname: profileEditView.RemoteHostname,
			Port:     profileEditView.RemoteSshPort,
			Auth: config.ProfileAuth{
				User:           profileEditView.RemoteSshUser,
				Password:       nil,
				PrivateKeyPath: nil,
			},
		}

		if profileEditView.RemoteSshPassword != "" {
			editedProfile.Auth.Password = &profileEditView.RemoteSshPassword
		} else if profileEditView.RemoteSshPrivateKeyPath != "" {
			editedProfile.Auth.PrivateKeyPath = &profileEditView.RemoteSshPrivateKeyPath
		} else {
			log.Fatal(errors.New("password or private key path must be provided"))
		}

		for i, profile := range c.Profiles {
			if profile.Id == chosenProfileId {
				c.Profiles[i] = editedProfile
				break
			}
		}

		err = c.Save()
		if err != nil {
			log.Fatal(err)
		}

		info_view.Render(info_view.ProfileInfo{
			ProfileName: profileEditView.ProfileName,
			ServerUrl:   profileEditView.RemoteHostname,
		}, "edited")
	},
}

func init() {
	profileEditCmd.Flags().StringVarP(&profileNameFlag, "name", "n", "", "Profile name")
	profileEditCmd.Flags().StringVarP(&serverHostnameFlag, "hostname", "h", "", "Remote SSH hostname")
	profileEditCmd.Flags().IntVarP(&serverPortFlag, "port", "P", 22, "Remote SSH port")
	profileEditCmd.Flags().StringVarP(&serverUserFlag, "user", "u", "", "Remote SSH url")
	profileEditCmd.Flags().StringVarP(&serverPasswordFlag, "password", "p", "", "Remote SSH password")
	profileEditCmd.Flags().StringVarP(&serverPrivateKeyPathFlag, "private-key-path", "k", "", "Remote SSH private key path")
}
