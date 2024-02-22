// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_server

import (
	"fmt"
	"strconv"
	"time"

	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/cli/remote_installer"
	"github.com/daytonaio/daytona/internal/util"

	cmd_profile "github.com/daytonaio/daytona/cli/cmd/profile"
	list_view "github.com/daytonaio/daytona/cli/cmd/views/profile/list_view"
	view "github.com/daytonaio/daytona/cli/cmd/views/server/installation_wizard"
	views_util "github.com/daytonaio/daytona/cli/cmd/views/util"

	"github.com/briandowns/spinner"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the Daytona Server on a remote host",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)

		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		profilesList := c.Profiles
		chosenProfileId := list_view.GetProfileIdFromPrompt(profilesList, c.ActiveProfileId, "Choose a profile to install on", true)

		if chosenProfileId == list_view.NewProfileId {
			chosenProfileId, err = cmd_profile.CreateProfile(c, nil, false, false)
			if err != nil {
				log.Fatal(err)
			}
		}

		chosenProfile, err := c.GetProfile(chosenProfileId)
		if err != nil {
			log.Fatal(err)
		}

		installDockerPrompt := true

		views_util.RenderMainTitle("REMOTE INSTALLER")

		var client *ssh.Client

		sshConfig := util.GetSshConfigFromProfile(&chosenProfile)

		fmt.Println("Connecting to the remote host ...")
		s.Start()
		defer s.Stop()

		client, err = ssh.Dial("tcp", chosenProfile.Hostname+":"+strconv.Itoa(chosenProfile.Port), sshConfig)
		if err != nil {
			log.Fatal(err)
		}

		installer := &remote_installer.RemoteInstaller{
			Client:     client,
			BinaryUrl:  config.GetBinaryUrls(),
			Downloader: remote_installer.DownloaderCurl,
		}

		s.Stop()

		remoteOs, err := installer.DetectOs()
		if err != nil {
			log.Error("Failed to detect remote OS")
			log.Fatal(err)
		}

		serverRegistered, err := installer.ServerRegistered()
		if err != nil {
			log.Fatal(err)
		}

		if serverRegistered {
			fmt.Println("Daytona Server is already installed on the remote machine.")
			return
		}

		sudoPasswordRequired, err := installer.SudoPasswordRequired()
		if err != nil {
			log.Fatal(err)
		}

		var sessionPassword string

		if sudoPasswordRequired && (chosenProfile.Auth.Password == nil || *chosenProfile.Auth.Password == "") {
			if chosenProfile.Auth.Password == nil || *chosenProfile.Auth.Password == "" {
				fmt.Printf("Enter password for user %s:", chosenProfile.Auth.User)
				password, err := term.ReadPassword(0)
				fmt.Println()
				if err != nil {
					log.Fatal(err)
				}
				sessionPassword = string(password)
			} else {
				sessionPassword = *chosenProfile.Auth.Password
			}
		}
		installer.Password = sessionPassword

		curlInstalled, err := installer.CurlInstalled()
		if err != nil {
			log.Fatal(err)
		}

		if !curlInstalled {
			wgetInstalled, err := installer.WgetInstalled()
			if err != nil {
				log.Fatal(err)
			}
			if !wgetInstalled {
				fmt.Println("Neither 'curl' nor 'wget' are installed on the remote machine. Please install one of them and try again.")
				return
			}
			installer.Downloader = remote_installer.DownloaderWget
		}

		dockerInstalled, err := installer.DockerInstalled()
		if err != nil {
			log.Fatal(err)
		}

		if !dockerInstalled {
			s.Stop()
			view.DockerPrompt(&installDockerPrompt)
			s.Start()
			if installDockerPrompt {

				fmt.Println("Installing Docker")
				s.Start()
				defer s.Stop()

				err := installer.InstallDocker(*remoteOs)
				if err != nil {
					log.Error("Failed to install Docker.")
					log.Fatal(err)
				}
			} else {
				log.Info("Installation terminated because Docker is required")
				return
			}
		}

		err = installer.AddUserToDockerGroup(chosenProfile.Auth.User)
		if err != nil {
			log.Fatal(err)
		}

		s.Stop()
		fmt.Println("Installing Daytona")
		s.Start()

		err = installer.InstallBinary(*remoteOs)
		if err != nil {
			log.Error("Failed to install Daytona binary")
			log.Fatal(err)
		}

		err = installer.RegisterDaemon(*remoteOs)
		if err != nil {
			log.Error("Failed to register daemon")
			log.Fatal(err)
		}

		err = installer.EnableServiceLinger(chosenProfile.Auth.User)
		if err != nil {
			log.Error("Failed to enable service linger")
			log.Fatal(err)
		}

		apiUrl, err := installer.GetApiUrl()
		if err != nil {
			log.Error("Failed to get API URL from the remote machine")
			log.Fatal(err)
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			log.Fatal(err)
		}

		activeProfile.Api.Url = apiUrl
		err = c.EditProfile(activeProfile)
		if err != nil {
			log.Error("Failed to set API URL from the remote machine")
			log.Fatal(err)
		}

		restartServerPrompt := true

		s.Stop()
		view.RestartPrompt(&restartServerPrompt)
		s.Start()
		if restartServerPrompt {

			fmt.Println("Restarting the remote machine")
			s.Start()
			defer s.Stop()

			installer.RestartServer()
		} else {
			log.Info("Please restart the remote machine manually")
			return
		}

		// Recreate the ssh client

		client, err = installer.WaitForRemoteServerToStart(chosenProfile.Hostname, chosenProfile.Port, sshConfig)
		if err != nil {
			log.Fatal(err)
		}

		installer.Client = client

		s.Stop()
		fmt.Println("Waiting for Daytona Server to start")
		s.Start()

		err = installer.WaitForDaytonaServerToStart(apiUrl)
		if err != nil {
			log.Error("Failed to wait for Daytona server to start")
			log.Fatal(err)
		}

		s.Stop()

		fmt.Println("\nDaytona Server has been successfully installed.")
		views_util.RenderInfoMessageBold("Use 'daytona create' to initialize your first workspace.")
	},
}
