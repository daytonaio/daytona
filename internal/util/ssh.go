// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path"

	"github.com/daytonaio/daytona/cli/config"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type SSHKeyPair struct {
	PublicKey  string
	PrivateKey string
}

func GetHostKey() (ssh.Signer, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	dir := path.Join(userConfigDir, "daytona", "ssh")

	hostKeyPath := path.Join(dir, "host_key")
	hostKey, err := os.ReadFile(hostKeyPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Generate a new host key
			privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			if err != nil {
				return nil, err
			}

			privateKeyDer := x509.MarshalPKCS1PrivateKey(privateKey)
			privateKeyBlock := pem.Block{
				Type:    "RSA PRIVATE KEY",
				Headers: nil,
				Bytes:   privateKeyDer,
			}
			privateKeyPem := pem.EncodeToMemory(&privateKeyBlock)

			err = os.MkdirAll(dir, 0755)
			if err != nil {
				return nil, err
			}

			// Write the new host key to the file
			err = os.WriteFile(hostKeyPath, privateKeyPem, 0600)
			if err != nil {
				return nil, err
			}

			hostKey = privateKeyPem
		} else {
			return nil, err
		}
	}

	signer, err := ssh.ParsePrivateKey(hostKey)
	if err != nil {
		return nil, err
	}

	return signer, nil
}

func GetSshConfigFromProfile(chosenProfile *config.Profile) *ssh.ClientConfig {
	sshConfig := &ssh.ClientConfig{
		User:            chosenProfile.Auth.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	if chosenProfile.Auth.Password != nil && *chosenProfile.Auth.Password != "" {
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.Password(*chosenProfile.Auth.Password),
		}
	} else if chosenProfile.Auth.PrivateKeyPath != nil && *chosenProfile.Auth.PrivateKeyPath != "" {

		privateKeyContent, err := os.ReadFile(*chosenProfile.Auth.PrivateKeyPath)
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
