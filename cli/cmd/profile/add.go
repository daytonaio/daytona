// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_profile

import (
	"errors"

	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/internal/util"

	view "github.com/daytonaio/daytona/cli/cmd/views/profile/creation_wizard"
	"github.com/daytonaio/daytona/cli/cmd/views/profile/info_view"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var profileAddCmd = &cobra.Command{
	Use:     "add",
	Short:   "Add profile",
	Args:    cobra.NoArgs,
	Aliases: []string{"new"},
	Run: func(cmd *cobra.Command, args []string) {
		profileAddView := view.ProfileAddView{
			ProfileName:             "",
			RemoteHostname:          "",
			RemoteSshPort:           22,
			RemoteSshPassword:       "",
			RemoteSshUser:           "",
			RemoteSshPrivateKeyPath: "",
		}

		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		if profileNameFlag == "" || serverHostnameFlag == "" || serverUserFlag == "" || (serverPrivateKeyPathFlag == "" && serverPasswordFlag == "") {
			view.ProfileCreationView(c, &profileAddView, false)
		} else {
			profileAddView.ProfileName = profileNameFlag
			profileAddView.RemoteHostname = serverHostnameFlag
			profileAddView.RemoteSshPassword = serverPasswordFlag
			profileAddView.RemoteSshUser = serverUserFlag
			profileAddView.RemoteSshPrivateKeyPath = serverPrivateKeyPathFlag
		}

		addProfile(profileAddView, c, true)
	},
}

func CreateProfile(c *config.Config, notify bool) string {
	profileAddView := view.ProfileAddView{
		ProfileName:             "",
		RemoteHostname:          "",
		RemoteSshPort:           22,
		RemoteSshPassword:       "",
		RemoteSshUser:           "",
		RemoteSshPrivateKeyPath: "",
	}

	view.ProfileCreationView(c, &profileAddView, false)
	return addProfile(profileAddView, c, notify)
}

func addProfile(profileAddView view.ProfileAddView, c *config.Config, notify bool) string {
	profile := config.Profile{
		Id:       util.GenerateIdFromName(profileAddView.ProfileName),
		Name:     profileAddView.ProfileName,
		Hostname: profileAddView.RemoteHostname,
		Port:     profileAddView.RemoteSshPort,
		Auth: config.ProfileAuth{
			User:           profileAddView.RemoteSshUser,
			Password:       nil,
			PrivateKeyPath: nil,
		},
	}

	if profileAddView.RemoteSshPassword != "" {
		profile.Auth.Password = &profileAddView.RemoteSshPassword
	} else if profileAddView.RemoteSshPrivateKeyPath != "" {
		profile.Auth.PrivateKeyPath = &profileAddView.RemoteSshPrivateKeyPath
	} else {
		log.Fatal(errors.New("password or private key path must be provided"))
	}

	err := c.AddProfile(profile)
	if err != nil {
		log.Fatal(err)
	}

	if notify {
		info_view.Render(info_view.ProfileInfo{
			ProfileName: profile.Name,
			ServerUrl:   profile.Hostname,
		}, "added and set as active")
	}

	return profile.Id
}

var profileNameFlag string
var serverHostnameFlag string
var serverPortFlag int
var serverPasswordFlag string
var serverUserFlag string
var serverPrivateKeyPathFlag string

func init() {
	profileAddCmd.Flags().StringVarP(&profileNameFlag, "name", "n", "", "Profile name")
	profileAddCmd.Flags().StringVarP(&serverHostnameFlag, "hostname", "h", "", "Remote hostname")
	profileAddCmd.Flags().IntVarP(&serverPortFlag, "port", "P", 22, "Remote SSH port")
	profileAddCmd.Flags().StringVarP(&serverUserFlag, "user", "u", "", "Remote SSH user")
	profileAddCmd.Flags().StringVarP(&serverPasswordFlag, "password", "p", "", "Remote SSH password")
	profileAddCmd.Flags().StringVarP(&serverPrivateKeyPathFlag, "private-key-path", "k", "", "Remote SSH private key path")
}
