// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package tailscale

import (
	"context"

	"github.com/daytonaio/daytona/pkg/tailscale/tunnel"
	log "github.com/sirupsen/logrus"
	"tailscale.com/tsnet"
)

type ForwardConfig struct {
	TsnetConn  *tsnet.Server
	Hostname   string
	SshPort    int
	LocalSock  string
	RemoteSock string
}

func ForwardRemoteUnixSock(ctx context.Context, config ForwardConfig) (chan bool, chan error) {
	sshTun := tunnel.NewUnix(config.TsnetConn, config.LocalSock, config.Hostname, config.SshPort, config.RemoteSock)

	errChan := make(chan error)

	sshTun.SetTunneledConnState(func(tun *tunnel.SshTunnel, state *tunnel.TunneledConnectionState) {
		log.Debugf("%+v", state)
	})

	startedChann := make(chan bool, 1)

	sshTun.SetConnState(func(tun *tunnel.SshTunnel, state tunnel.ConnectionState) {
		switch state {
		case tunnel.StateStarting:
			log.Debugf("SSH Tunnel is Starting")
		case tunnel.StateStarted:
			log.Debugf("SSH Tunnel is Started")
			startedChann <- true
		case tunnel.StateStopped:
			log.Debugf("SSH Tunnel is Stopped")
		}
	})

	go func() {
		errChan <- sshTun.Start(ctx)
	}()

	return startedChann, errChan
}
