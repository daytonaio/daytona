// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ssh

import (
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

type SessionConfig struct {
	Hostname       string
	Port           int
	Username       string
	Password       *string
	PrivateKeyPath *string
}

type Client struct {
	*ssh.Client
}

func NewClient(config *SessionConfig) (*Client, error) {
	server := fmt.Sprintf("%s:%d", config.Hostname, config.Port)
	sshConfig := &ssh.ClientConfig{
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: time.Second * 30,
		User:    config.Username,
	}

	auth := []ssh.AuthMethod{}

	if config.Password != nil {
		auth = append(auth, []ssh.AuthMethod{
			ssh.Password(*config.Password),
		}...)
	}

	if config.PrivateKeyPath != nil {
		buf, err := os.ReadFile(*config.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("reading SSH key file %s: %w", *config.PrivateKeyPath, err)
		}

		privateKey, err := ssh.ParsePrivateKey(buf)
		if err != nil {
			return nil, err
		}

		auth = append(auth, []ssh.AuthMethod{
			ssh.PublicKeys(privateKey),
		}...)
	}

	sshConfig.Auth = auth

	client, err := ssh.Dial("tcp", server, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("dialing SSH server: %w", err)
	}

	return &Client{
		Client: client,
	}, nil
}
