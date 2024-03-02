// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"os"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

func GetSshConfigFromProfile(chosenProfile *config.Profile) *ssh.ClientConfig {
	sshConfig := &ssh.ClientConfig{
		User:            chosenProfile.RemoteAuth.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	if chosenProfile.RemoteAuth.Password != nil && *chosenProfile.RemoteAuth.Password != "" {
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.Password(*chosenProfile.RemoteAuth.Password),
		}
	} else if chosenProfile.RemoteAuth.PrivateKeyPath != nil && *chosenProfile.RemoteAuth.PrivateKeyPath != "" {

		privateKeyContent, err := os.ReadFile(*chosenProfile.RemoteAuth.PrivateKeyPath)
		if err != nil {
			log.Fatal(err)
		}

		privateKey, err := ssh.ParsePrivateKey([]byte(privateKeyContent))
		if err != nil {
			log.Fatal(err)
		}

		sshConfig.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(privateKey),
		}
	} else {
		log.Fatal("No authentication method provided")
	}

	return sshConfig
}
