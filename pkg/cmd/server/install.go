// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_server

import (
	"fmt"
	"strconv"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	cmd_profile "github.com/daytonaio/daytona/pkg/cmd/profile"
	"github.com/daytonaio/daytona/pkg/remoteinstaller"
	"github.com/daytonaio/daytona/pkg/views/profile"
	"github.com/daytonaio/daytona/pkg/views/server"
	view_util "github.com/daytonaio/daytona/pkg/views/util"

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
		chosenProfileId := profile.GetProfileIdFromPrompt(profilesList, c.ActiveProfileId, "Choose a profile to install on", true)

		if chosenProfileId == profile.NewProfileId {
			chosenProfileId, err = cmd_profile.CreateProfile(c, nil, false, false, true)
			if err != nil {
				log.Fatal(err)
			}
		}

		chosenProfile, err := c.GetProfile(chosenProfileId)
		if err != nil {
			log.Fatal(err)
		}

		if chosenProfile.RemoteAuth == nil {
			err = cmd_profile.EditProfile(c, false, &chosenProfile, true, true)
			if err != nil {
				log.Fatal(err)
			}
		}

		installDockerPrompt := true

		view_util.RenderMainTitle("REMOTE INSTALLER")

		var client *ssh.Client

		sshConfig := util.GetSshConfigFromProfile(&chosenProfile)

		fmt.Println("Connecting to the remote host ...")
		s.Start()
		defer s.Stop()

		client, err = ssh.Dial("tcp", chosenProfile.RemoteAuth.Hostname+":"+strconv.Itoa(chosenProfile.RemoteAuth.Port), sshConfig)
		if err != nil {
			log.Fatal(err)
		}

		installer := &remoteinstaller.RemoteInstaller{
			Client:     client,
			BinaryUrl:  config.GetBinaryUrls(),
			Downloader: remoteinstaller.DownloaderCurl,
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

		if sudoPasswordRequired && (chosenProfile.RemoteAuth.Password == nil || *chosenProfile.RemoteAuth.Password == "") {
			if chosenProfile.RemoteAuth.Password == nil || *chosenProfile.RemoteAuth.Password == "" {
				fmt.Printf("Enter password for user %s:", chosenProfile.RemoteAuth.User)
				password, err := term.ReadPassword(0)
				fmt.Println()
				if err != nil {
					log.Fatal(err)
				}
				sessionPassword = string(password)
			} else {
				sessionPassword = *chosenProfile.RemoteAuth.Password
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
			installer.Downloader = remoteinstaller.DownloaderWget
		}

		dockerInstalled, err := installer.DockerInstalled()
		if err != nil {
			log.Fatal(err)
		}

		if !dockerInstalled {
			s.Stop()
			server.DockerPrompt(&installDockerPrompt)
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

		err = installer.AddUserToDockerGroup(chosenProfile.RemoteAuth.User)
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

		err = installer.EnableServiceLinger(chosenProfile.RemoteAuth.User)
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
		server.RestartPrompt(&restartServerPrompt)
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

		client, err = installer.WaitForRemoteServerToStart(chosenProfile.RemoteAuth.Hostname, chosenProfile.RemoteAuth.Port, sshConfig)
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
		view_util.RenderInfoMessageBold("Use 'daytona create' to initialize your first workspace.")
	},
}
