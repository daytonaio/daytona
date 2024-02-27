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
	view_util "github.com/daytonaio/daytona/pkg/views/util"

	"github.com/briandowns/spinner"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the Daytona Server from a remote host",
	Long:  "Uninstall the Daytona Server from a remote host. Note: this command will not uninstall Docker from your system",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)

		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		profilesList := c.Profiles
		chosenProfileId := profile.GetProfileIdFromPrompt(profilesList, c.ActiveProfileId, "Choose a profile to uninstall from", true)

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

		view_util.RenderMainTitle("REMOTE UNINSTALLER")

		var client *ssh.Client

		sshConfig := util.GetSshConfigFromProfile(&chosenProfile)

		fmt.Println("Connecting to remote host ...")
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

		if !serverRegistered {
			fmt.Println("Daytona Server is not installed on the remote machine.\n")
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

		err = installer.RemoveDaemon(*remoteOs)
		if err != nil {
			log.Error("Failed to remove Daytona daemon")
			log.Fatal(err)
		}

		err = installer.RemoveBinary(*remoteOs)
		if err != nil {
			log.Error("Failed to remove Daytona binary")
			log.Fatal(err)
		}

		fmt.Println("\nDaytona Server has been successfully uninstalled.\n")
	},
}
