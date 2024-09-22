// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package tunnel

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/errgroup"
	"tailscale.com/tsnet"
)

type SshTunnel struct {
	mutex             *sync.Mutex
	ctx               context.Context
	cancel            context.CancelFunc
	started           bool
	server            *Endpoint
	local             *Endpoint
	remote            *Endpoint
	timeout           time.Duration
	connState         func(*SshTunnel, ConnectionState)
	tunneledConnState func(*SshTunnel, *TunneledConnectionState)
	active            int
	sshClient         *ssh.Client
	sshConfig         *ssh.ClientConfig
	tsnetConn         *tsnet.Server
}

// ConnectionState represents the state of the SSH tunnel. It's returned to an optional function provided to SetConnState.
type ConnectionState int

const (
	// StateStopped represents a stopped tunnel. A call to Start will make the state to transition to StateStarting.
	StateStopped ConnectionState = iota

	// StateStarting represents a tunnel initializing and preparing to listen for connections.
	// A successful initialization will make the state to transition to StateStarted, otherwise it will transition to StateStopped.
	StateStarting

	// StateStarted represents a tunnel ready to accept connections.
	// A call to stop or an error will make the state to transition to StateStopped.
	StateStarted
)

// New creates a new SSH tunnel to the specified server redirecting a port on local localhost to a port on remote localhost.
// The states of the tunnel can be received through a callback function with SetConnState.
// The states of the tunneled connections can be received through a callback function with SetTunneledConnState.
func New(tsnetConn *tsnet.Server, localPort int, server string, serverPort, remotePort int) *SshTunnel {
	sshTun := defaultSSHTun(server, serverPort, tsnetConn)
	sshTun.local = NewTCPEndpoint("localhost", localPort)
	sshTun.remote = NewTCPEndpoint("localhost", remotePort)
	return sshTun
}

// NewUnix does the same as New but using unix sockets.
func NewUnix(tsnetConn *tsnet.Server, localUnixSocket, server string, serverPort int, remoteUnixSocket string) *SshTunnel {
	sshTun := defaultSSHTun(server, serverPort, tsnetConn)
	sshTun.local = NewUnixEndpoint(localUnixSocket)
	sshTun.remote = NewUnixEndpoint(remoteUnixSocket)
	return sshTun
}

func defaultSSHTun(server string, port int, tsnetConn *tsnet.Server) *SshTunnel {
	return &SshTunnel{
		mutex:     &sync.Mutex{},
		server:    NewTCPEndpoint(server, port),
		timeout:   time.Second * 15,
		tsnetConn: tsnetConn,
	}
}

// SetConnState specifies an optional callback function that is called when a SSH tunnel changes state.
// See the ConnState type and associated constants for details.
func (tun *SshTunnel) SetConnState(connStateFun func(*SshTunnel, ConnectionState)) {
	tun.connState = connStateFun
}

// SetTunneledConnState specifies an optional callback function that is called when the underlying tunneled
// connections change state.
func (tun *SshTunnel) SetTunneledConnState(tunneledConnStateFun func(*SshTunnel, *TunneledConnectionState)) {
	tun.tunneledConnState = tunneledConnStateFun
}

// Start starts the SSH tunnel. It can be stopped by calling `Stop` or cancelling its context.
// This call will block until the tunnel is stopped either calling those methods or by an error.
func (tun *SshTunnel) Start(ctx context.Context) error {
	tun.mutex.Lock()
	if tun.started {
		tun.mutex.Unlock()
		return errors.New("already started")
	}
	tun.started = true
	tun.ctx, tun.cancel = context.WithCancel(ctx)
	tun.mutex.Unlock()

	if tun.connState != nil {
		tun.connState(tun, StateStarting)
	}

	config, err := tun.initSSHConfig()
	if err != nil {
		return tun.stop(fmt.Errorf("ssh config failed: %w", err))
	}
	tun.sshConfig = config

	listenConfig := net.ListenConfig{}
	localListener, err := listenConfig.Listen(tun.ctx, tun.local.Type(), tun.local.String())
	if err != nil {
		return tun.stop(fmt.Errorf("local listen %s on %s failed: %w", tun.local.Type(), tun.local.String(), err))
	}

	errChan := make(chan error)
	go func() {
		errChan <- tun.listen(localListener)
	}()

	if tun.connState != nil {
		tun.connState(tun, StateStarted)
	}

	return tun.stop(<-errChan)
}

// Stop closes all connections and makes Start exit gracefuly.
func (tun *SshTunnel) Stop() {
	tun.mutex.Lock()
	defer tun.mutex.Unlock()

	if tun.started {
		tun.cancel()
	}
}

func (tun *SshTunnel) initSSHConfig() (*ssh.ClientConfig, error) {
	config := &ssh.ClientConfig{
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: tun.timeout,
	}

	return config, nil
}

func (tun *SshTunnel) stop(err error) error {
	tun.mutex.Lock()
	tun.started = false
	tun.mutex.Unlock()
	if tun.connState != nil {
		tun.connState(tun, StateStopped)
	}
	return err
}

func (tun *SshTunnel) listen(localListener net.Listener) error {
	errGroup, groupCtx := errgroup.WithContext(tun.ctx)

	errGroup.Go(func() error {
		for {
			localConn, err := localListener.Accept()
			if err != nil {
				return fmt.Errorf("local accept %s on %s failed: %w", tun.local.Type(), tun.local.String(), err)
			}

			errGroup.Go(func() error {
				return tun.handle(localConn)
			})
		}
	})

	<-groupCtx.Done()

	localListener.Close()

	err := errGroup.Wait()

	select {
	case <-tun.ctx.Done():
	default:
		return err
	}

	return nil
}

func (tun *SshTunnel) handle(localConn net.Conn) error {
	err := tun.addConn()
	if err != nil {
		return err
	}

	tun.forward(localConn)
	tun.removeConn()

	return nil
}

func (tun *SshTunnel) addConn() error {
	tun.mutex.Lock()
	defer tun.mutex.Unlock()

	if tun.active == 0 {
		if tun.tsnetConn == nil {
			return errors.New("tsnetConn is not set")
		}
		conn, err := tun.tsnetConn.Dial(tun.ctx, tun.server.Type(), tun.server.String())
		if err != nil {
			return err
		}
		c, chans, reqs, err := ssh.NewClientConn(conn, tun.server.String(), tun.sshConfig)
		if err != nil {
			return err
		}
		tun.sshClient = ssh.NewClient(c, chans, reqs)
	}

	tun.active += 1

	return nil
}

func (tun *SshTunnel) removeConn() {
	tun.mutex.Lock()
	defer tun.mutex.Unlock()

	tun.active -= 1

	if tun.active == 0 {
		tun.sshClient.Close()
		tun.sshClient = nil
	}
}
