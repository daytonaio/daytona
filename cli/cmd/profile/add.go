// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_profile

import (
	"errors"
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

		if profileNameFlag == "" || serverHostnameFlag == "" || serverUserFlag == "" || provisionerFlag == "" || (serverPrivateKeyPathFlag == "" && serverPasswordFlag == "") {
			_, err = CreateProfile(c, nil, true, true)
			if err != nil {
				log.Fatal(err)
			}
			return
		}

		profileAddView := view.ProfileAddView{
			ProfileName:    profileNameFlag,
			RemoteHostname: serverHostnameFlag,
			RemoteSshPort:  serverPortFlag,
			RemoteSshUser:  serverUserFlag,
			Provisioner:    provisionerFlag,
		}
		if serverPasswordFlag != "" {
			profileAddView.RemoteSshPassword = serverPasswordFlag
		} else if serverPrivateKeyPathFlag != "" {
			profileAddView.RemoteSshPrivateKeyPath = serverPrivateKeyPathFlag
		} else {
			log.Fatal(errors.New("password or private key path must be provided"))
		}

		addProfile(profileAddView, c, true, true)
	},
}

func CreateProfile(c *config.Config, profileAddView *view.ProfileAddView, checkConnection bool, notify bool) (string, error) {
	if profileAddView == nil {
		profileAddView = &view.ProfileAddView{
			ProfileName:             "",
			RemoteHostname:          "",
			RemoteSshPort:           22,
			RemoteSshPassword:       "",
			RemoteSshUser:           "",
			RemoteSshPrivateKeyPath: "",
			Provisioner:             "",
		}
	}

	view.ProfileCreationView(c, profileAddView, false)

	return addProfile(*profileAddView, c, checkConnection, notify)
}

func addProfile(profileView view.ProfileAddView, c *config.Config, checkConnection bool, notify bool) (string, error) {
	profile := config.Profile{
		Id:       util.GenerateIdFromName(profileView.ProfileName),
		Name:     profileView.ProfileName,
		Hostname: profileView.RemoteHostname,
		Port:     profileView.RemoteSshPort,
		Auth: config.ProfileAuth{
			User:           profileView.RemoteSshUser,
			Password:       nil,
			PrivateKeyPath: nil,
		},
	}

	if profileView.RemoteSshPassword != "" {
		profile.Auth.Password = &profileView.RemoteSshPassword
	} else if profileView.RemoteSshPrivateKeyPath != "" {
		profile.Auth.PrivateKeyPath = &profileView.RemoteSshPrivateKeyPath
	} else {
		return "", errors.New("password or private key path must be provided")
	}

	if checkConnection {
		ignoreConnectionCheckPrompt := false
		err := checkDaytonaInstalled(profileView, profile)
		if err != nil {
			view.IgnoreConnectionFailedCheck(&ignoreConnectionCheckPrompt, err.Error())
			if !ignoreConnectionCheckPrompt {
				view.ProfileCreationView(c, &profileView, false)

				return addProfile(profileView, c, true, notify)
			}
		}
	}

	err := c.AddProfile(profile)
	if err != nil {
		return "", err
	}

	if notify {
		info_view.Render(info_view.ProfileInfo{
			ProfileName: profile.Name,
			ServerUrl:   profile.Hostname,
		}, "added and set as active")
	}

	return profile.Id, nil
}

func checkDaytonaInstalled(profileView view.ProfileAddView, profile config.Profile) error {

	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)

	sshConfig := util.GetSshConfigFromProfile(&profile)

	s.Start()
	defer s.Stop()

	client, err := ssh.Dial("tcp", profile.Hostname+":"+strconv.Itoa(profile.Port), sshConfig)
	if err != nil {
		return errors.New("Failed to connect to the remote machine")
	}

	installer := &remote_installer.RemoteInstaller{
		Client: client,
	}

	s.Stop()

	serverRegistered, err := installer.ServerRegistered()
	if err != nil {
		return errors.New("Failed to execute command on remote machine")
	}

	if !serverRegistered {
		return errors.New("Daytona Server is not running on the remote machine")
	}
	return nil
}

var profileNameFlag string
var serverHostnameFlag string
var serverPortFlag int = 0
var serverPasswordFlag string
var serverUserFlag string
var serverPrivateKeyPathFlag string
var provisionerFlag string

func init() {
	profileAddCmd.Flags().StringVarP(&profileNameFlag, "name", "n", "", "Profile name")
	profileAddCmd.Flags().StringVarP(&serverHostnameFlag, "hostname", "h", "", "Remote hostname")
	profileAddCmd.Flags().IntVarP(&serverPortFlag, "port", "P", 22, "Remote SSH port")
	profileAddCmd.Flags().StringVarP(&serverUserFlag, "user", "u", "", "Remote SSH user")
	profileAddCmd.Flags().StringVarP(&serverPasswordFlag, "password", "p", "", "Remote SSH password")
	profileAddCmd.Flags().StringVarP(&serverPrivateKeyPathFlag, "private-key-path", "k", "", "Remote SSH private key path")
	profileAddCmd.Flags().StringVarP(&provisionerFlag, "provisioner", "r", "default", "Provisioner")
}
