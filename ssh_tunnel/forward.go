// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ssh_tunnel

import (
	"context"
	"fmt"
	"io"
	"net"

	"golang.org/x/sync/errgroup"
)

// TunneledConnectionState represents the state of the final connections made through the tunnel.
type TunneledConnectionState struct {
	// From is the address initating the connection.
	From string
	// Info holds a message with info on the state of the connection (useful for debug purposes).
	Info string
	// Error holds an error on the connection or nil if the connection is successful.
	Error error
	// Ready indicates if the connection is established.
	Ready bool
	// Closed indicates if the coonnection is closed.
	Closed bool
}

func (s *TunneledConnectionState) String() string {
	out := fmt.Sprintf("[%s] ", s.From)
	if s.Info != "" {
		out += s.Info
	}
	if s.Error != nil {
		out += fmt.Sprintf("Error: %v", s.Error)
	}
	return out
}

func (tun *SshTunnel) forward(localConn net.Conn) {
	from := localConn.RemoteAddr().String()

	tun.tunneledState(&TunneledConnectionState{
		From: from,
		Info: fmt.Sprintf("accepted %s connection", tun.local.Type()),
	})

	remoteConn, err := tun.sshClient.Dial(tun.remote.Type(), tun.remote.String())
	if err != nil {
		tun.tunneledState(&TunneledConnectionState{
			From:  from,
			Error: fmt.Errorf("remote dial %s to %s failed: %w", tun.remote.Type(), tun.remote.String(), err),
		})

		localConn.Close()
		return
	}

	connStr := fmt.Sprintf("%s -(%s)> %s -(ssh)> %s -(%s)> %s", from, tun.local.Type(), tun.local.String(),
		tun.server.String(), tun.remote.Type(), tun.remote.String())

	tun.tunneledState(&TunneledConnectionState{
		From:   from,
		Info:   fmt.Sprintf("connection established: %s", connStr),
		Ready:  true,
		Closed: false,
	})

	connCtx, connCancel := context.WithCancel(tun.ctx)
	errGroup := &errgroup.Group{}

	errGroup.Go(func() error {
		defer connCancel()
		_, err = io.Copy(remoteConn, localConn)
		if err != nil {
			return fmt.Errorf("failed copying bytes from remote to local: %w", err)
		}
		return remoteConn.Close()
	})

	errGroup.Go(func() error {
		defer connCancel()
		_, err = io.Copy(localConn, remoteConn)
		if err != nil {
			return fmt.Errorf("failed copying bytes from local to remote: %w", err)
		}
		return localConn.Close()
	})

	err = errGroup.Wait()

	<-connCtx.Done()

	select {
	case <-tun.ctx.Done():
	default:
		if err != nil {
			tun.tunneledState(&TunneledConnectionState{
				From:  from,
				Error: err,
			})
		}
	}

	tun.tunneledState(&TunneledConnectionState{
		From:   from,
		Info:   fmt.Sprintf("connection closed: %s", connStr),
		Ready:  false,
		Closed: true,
	})
}

func (tun *SshTunnel) tunneledState(state *TunneledConnectionState) {
	if tun.tunneledConnState != nil {
		tun.tunneledConnState(tun, state)
	}
}
