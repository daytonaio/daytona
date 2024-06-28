// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package tailscale

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
	"tailscale.com/tsnet"
)

type ExecConfig struct {
	Ctx       context.Context
	TsnetConn *tsnet.Server
	Hostname  string
	SshPort   int
	LogWriter io.Writer
	Command   string
}

func ExecCommand(config ExecConfig) error {
	server := fmt.Sprintf("%s:%d", config.Hostname, config.SshPort)
	sshConfig := &ssh.ClientConfig{
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: time.Second * 30,
	}

	conn, err := config.TsnetConn.Dial(config.Ctx, "tcp", server)
	if err != nil {
		return err
	}
	c, chans, reqs, err := ssh.NewClientConn(conn, server, sshConfig)
	if err != nil {
		return err
	}

	sshClient := ssh.NewClient(c, chans, reqs)
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	if err != nil {
		return err
	}

	session.Stdout = config.LogWriter
	session.Stderr = config.LogWriter

	return session.Run(config.Command)
}
