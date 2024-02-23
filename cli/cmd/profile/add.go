// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_profile

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/briandowns/spinner"
	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/cli/remote_installer"
	"github.com/daytonaio/daytona/internal/util"
	"golang.org/x/crypto/ssh"

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
		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		profileAddView := view.ProfileAddView{
			ProfileName:        profileNameFlag,
			RemoteHostname:     serverHostnameFlag,
			RemoteSshPort:      serverPortFlag,
			RemoteSshUser:      serverUserFlag,
			DefaultProvisioner: provisionerFlag,
			ApiUrl:             apiUrlFlag,
		}

		if profileAddView.ProfileName != "" && profileAddView.ApiUrl != "" {
			_, err = addProfile(profileAddView, c, true, true)
		}

		if profileNameFlag == "" || serverHostnameFlag == "" || serverUserFlag == "" || provisionerFlag == "" || (serverPrivateKeyPathFlag == "" && serverPasswordFlag == "") {
			_, err = CreateProfile(c, nil, true, true, false)
		}

		if err != nil {
			log.Fatal(err)
		}
	},
}

func CreateProfile(c *config.Config, profileAddView *view.ProfileAddView, checkConnection, notify, forceRemoteAccess bool) (string, error) {
	if profileAddView == nil {
		profileAddView = &view.ProfileAddView{
			ProfileName:             "",
			RemoteHostname:          "",
			RemoteSshPort:           22,
			RemoteSshPassword:       "",
			RemoteSshUser:           "",
			RemoteSshPrivateKeyPath: "",
			DefaultProvisioner:      "",
			ApiUrl:                  "",
		}
	}

	view.ProfileCreationView(c, profileAddView, false, forceRemoteAccess, false)

	return addProfile(*profileAddView, c, checkConnection, notify)
}

func addProfile(profileView view.ProfileAddView, c *config.Config, checkConnection bool, notify bool) (string, error) {
	profile := config.Profile{
		Id:   util.GenerateIdFromName(profileView.ProfileName),
		Name: profileView.ProfileName,
		Api: config.ServerApi{
			Url: profileView.ApiUrl,
		},
	}

	if profile.Api.Url == "" {
		profile.RemoteAuth = &config.RemoteAuth{
			Port:           profileView.RemoteSshPort,
			Hostname:       profileView.RemoteHostname,
			User:           profileView.RemoteSshUser,
			Password:       nil,
			PrivateKeyPath: nil,
		}

		if profileView.RemoteSshPassword != "" {
			profile.RemoteAuth.Password = &profileView.RemoteSshPassword
		} else if profileView.RemoteSshPrivateKeyPath != "" {
			profile.RemoteAuth.PrivateKeyPath = &profileView.RemoteSshPrivateKeyPath
		} else {
			return "", errors.New("password or private key path must be provided")
		}
	}

	if checkConnection && profile.Api.Url == "" {
		ignoreConnectionCheckPrompt := false
		err := setDaytonaApiUrl(&profileView, profile)
		if err != nil {
			fmt.Println(err.Error())
			view.IgnoreConnectionFailedCheck(&ignoreConnectionCheckPrompt, err.Error())
			if !ignoreConnectionCheckPrompt {
				view.ProfileCreationView(c, &profileView, false, false, false)

				return addProfile(profileView, c, true, notify)
			}
		}
	}

	profile.Api.Url = profileView.ApiUrl
	err := c.AddProfile(profile)
	if err != nil {
		return "", err
	}

	if notify {
		info_view.Render(info_view.ProfileInfo{
			ProfileName: profile.Name,
			ApiUrl:      profile.Api.Url,
		}, "added and set as active")
	}

	return profile.Id, nil
}

func setDaytonaApiUrl(profileView *view.ProfileAddView, profile config.Profile) error {
	if profile.RemoteAuth == nil {
		return errors.New("RemoteAuth is not set")
	}

	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)

	sshConfig := util.GetSshConfigFromProfile(&profile)

	s.Start()
	defer s.Stop()

	client, err := ssh.Dial("tcp", profile.RemoteAuth.Hostname+":"+strconv.Itoa(profile.RemoteAuth.Port), sshConfig)
	if err != nil {
		return errors.New("failed to connect to the remote machine")
	}

	installer := &remote_installer.RemoteInstaller{
		Client: client,
	}

	s.Stop()

	apiUrl, err := installer.GetApiUrl()
	if err != nil {
		return errors.New("failed to execute command on remote machine")
	}

	if apiUrl == "" {
		return errors.New("Daytona Server is not running on the remote machine")
	}

	profileView.ApiUrl = apiUrl
	return nil
}

var profileNameFlag string
var serverHostnameFlag string
var serverPortFlag int = 0
var serverPasswordFlag string
var serverUserFlag string
var serverPrivateKeyPathFlag string
var provisionerFlag string
var apiUrlFlag string

func init() {
	profileAddCmd.Flags().StringVarP(&profileNameFlag, "name", "n", "", "Profile name")
	profileAddCmd.Flags().StringVarP(&serverHostnameFlag, "hostname", "h", "", "Remote hostname")
	profileAddCmd.Flags().IntVarP(&serverPortFlag, "port", "P", 22, "Remote SSH port")
	profileAddCmd.Flags().StringVarP(&serverUserFlag, "user", "u", "", "Remote SSH user")
	profileAddCmd.Flags().StringVarP(&serverPasswordFlag, "password", "p", "", "Remote SSH password")
	profileAddCmd.Flags().StringVarP(&serverPrivateKeyPathFlag, "private-key-path", "k", "", "Remote SSH private key path")
	profileAddCmd.Flags().StringVarP(&provisionerFlag, "provisioner", "r", "default", "Provisioner")
	profileAddCmd.Flags().StringVarP(&apiUrlFlag, "api-url", "a", "", "API URL")
}
