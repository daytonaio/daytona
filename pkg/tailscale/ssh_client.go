// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package tailscale

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/daytonaio/daytona/pkg/ssh"
	cssh "golang.org/x/crypto/ssh"
	"tailscale.com/tsnet"
)

func NewSshClient(tsnetConn *tsnet.Server, sessionConfig *ssh.SessionConfig) (*ssh.Client, error) {
	server := fmt.Sprintf("%s:%d", sessionConfig.Hostname, sessionConfig.Port)
	conn, err := tsnetConn.Dial(context.Background(), "tcp", server)
	if err != nil {
		return nil, err
	}

	c, chans, reqs, err := cssh.NewClientConn(conn, server, &cssh.ClientConfig{
		HostKeyCallback: func(hostname string, remote net.Addr, key cssh.PublicKey) error {
			return nil
		},
		Timeout: 10 * time.Minute,
	})

	if err != nil {
		return nil, err
	}

	return &ssh.Client{
		Client: cssh.NewClient(c, chans, reqs),
	}, nil
}
