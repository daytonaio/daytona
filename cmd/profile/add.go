// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_profile

import (
	"errors"

	"github.com/daytonaio/daytona/config"
	"github.com/daytonaio/daytona/internal/util"

	view "github.com/daytonaio/daytona/cmd/views/profile_create_wizard"
	"github.com/daytonaio/daytona/cmd/views/profile_info"

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

		if profileNameFlag == "" || agentHostnameFlag == "" || agentUserFlag == "" || (agentPrivateKeyPathFlag == "" && agentPasswordFlag == "") {
			view.ProfileCreationView(c, &profileAddView, false)
		} else {
			profileAddView.ProfileName = profileNameFlag
			profileAddView.RemoteHostname = agentHostnameFlag
			profileAddView.RemoteSshPassword = agentPasswordFlag
			profileAddView.RemoteSshUser = agentUserFlag
			profileAddView.RemoteSshPrivateKeyPath = agentPrivateKeyPathFlag
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
		profile_info.Render(profile_info.ProfileInfo{
			ProfileName: profile.Name,
			AgentUrl:    profile.Hostname,
		}, "added and set as active")
	}

	return profile.Id
}

var profileNameFlag string
var agentHostnameFlag string
var agentPortFlag int
var agentPasswordFlag string
var agentUserFlag string
var agentPrivateKeyPathFlag string

func init() {
	profileAddCmd.Flags().StringVarP(&profileNameFlag, "name", "n", "", "Profile name")
	profileAddCmd.Flags().StringVarP(&agentHostnameFlag, "hostname", "h", "", "Remote hostname")
	profileAddCmd.Flags().IntVarP(&agentPortFlag, "port", "P", 22, "Remote SSH port")
	profileAddCmd.Flags().StringVarP(&agentUserFlag, "user", "u", "", "Remote SSH user")
	profileAddCmd.Flags().StringVarP(&agentPasswordFlag, "password", "p", "", "Remote SSH password")
	profileAddCmd.Flags().StringVarP(&agentPrivateKeyPathFlag, "private-key-path", "k", "", "Remote SSH private key path")
}
