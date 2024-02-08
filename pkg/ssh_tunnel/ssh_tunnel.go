// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ssh_tunnel

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/errgroup"
)

// SshTunnel represents a SSH tunnel
type SshTunnel struct {
	mutex             *sync.Mutex
	ctx               context.Context
	cancel            context.CancelFunc
	started           bool
	user              string
	authType          AuthType
	authKeyFile       string
	authKeyReader     io.Reader
	authPassword      string
	server            *Endpoint
	local             *Endpoint
	remote            *Endpoint
	timeout           time.Duration
	connState         func(*SshTunnel, ConnectionState)
	tunneledConnState func(*SshTunnel, *TunneledConnectionState)
	active            int
	sshClient         *ssh.Client
	sshConfig         *ssh.ClientConfig
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
// By default the SSH connection is made to port 22 as root and using automatic detection of the authentication
// method (see Start for details on this).
// Calling SetPassword will change the authentication to password based.
// Calling SetKeyFile will change the authentication to keyfile based..
// The SSH user and port can be changed with SetUser and SetPort.
// The local and remote hosts can be changed to something different than localhost with SetLocalEndpoint and SetRemoteEndpoint.
// The states of the tunnel can be received through a callback function with SetConnState.
// The states of the tunneled connections can be received through a callback function with SetTunneledConnState.
func New(localPort int, server string, remotePort int) *SshTunnel {
	sshTun := defaultSSHTun(server)
	sshTun.local = NewTCPEndpoint("localhost", localPort)
	sshTun.remote = NewTCPEndpoint("localhost", remotePort)
	return sshTun
}

// NewUnix does the same as New but using unix sockets.
func NewUnix(localUnixSocket string, server string, remoteUnixSocket string) *SshTunnel {
	sshTun := defaultSSHTun(server)
	sshTun.local = NewUnixEndpoint(localUnixSocket)
	sshTun.remote = NewUnixEndpoint(remoteUnixSocket)
	return sshTun
}

func defaultSSHTun(server string) *SshTunnel {
	return &SshTunnel{
		mutex:    &sync.Mutex{},
		server:   NewTCPEndpoint(server, 22),
		user:     "root",
		authType: AuthTypeAuto,
		timeout:  time.Second * 15,
	}
}

// SetPort changes the port where the SSH connection will be made.
func (tun *SshTunnel) SetPort(port int) {
	tun.server.port = port
}

// SetUser changes the user used to make the SSH connection.
func (tun *SshTunnel) SetUser(user string) {
	tun.user = user
}

// SetKeyFile changes the authentication to key-based and uses the specified file.
// Leaving the file empty defaults to the default Linux private key locations: `~/.ssh/id_rsa`, `~/.ssh/id_dsa`,
// `~/.ssh/id_ecdsa`, `~/.ssh/id_ecdsa_sk`, `~/.ssh/id_ed25519` and `~/.ssh/id_ed25519_sk`.
func (tun *SshTunnel) SetKeyFile(file string) {
	tun.authType = AuthTypeKeyFile
	tun.authKeyFile = file
}

// SetEncryptedKeyFile changes the authentication to encrypted key-based and uses the specified file and password.
// Leaving the file empty defaults to the default Linux private key locations: `~/.ssh/id_rsa`, `~/.ssh/id_dsa`,
// `~/.ssh/id_ecdsa`, `~/.ssh/id_ecdsa_sk`, `~/.ssh/id_ed25519` and `~/.ssh/id_ed25519_sk`.
func (tun *SshTunnel) SetEncryptedKeyFile(file string, password string) {
	tun.authType = AuthTypeEncryptedKeyFile
	tun.authKeyFile = file
	tun.authPassword = password
}

// SetKeyReader changes the authentication to key-based and uses the specified reader.
func (tun *SshTunnel) SetKeyReader(reader io.Reader) {
	tun.authType = AuthTypeKeyReader
	tun.authKeyReader = reader
}

// SetEncryptedKeyReader changes the authentication to encrypted key-based and uses the specified reader and password.
func (tun *SshTunnel) SetEncryptedKeyReader(reader io.Reader, password string) {
	tun.authType = AuthTypeEncryptedKeyReader
	tun.authKeyReader = reader
	tun.authPassword = password
}

// SetSSHServer changes the authentication to ssh-server.
func (tun *SshTunnel) SetSSHServer() {
	tun.authType = AuthTypeSSHServer
}

// SetPassword changes the authentication to password-based and uses the specified password.
func (tun *SshTunnel) SetPassword(password string) {
	tun.authType = AuthTypePassword
	tun.authPassword = password
}

// SetLocalHost sets the local host to redirect (defaults to localhost).
func (tun *SshTunnel) SetLocalHost(host string) {
	tun.local.host = host
}

// SetRemoteHost sets the remote host to redirect (defaults to localhost).
func (tun *SshTunnel) SetRemoteHost(host string) {
	tun.remote.host = host
}

// SetLocalEndpoint sets the local endpoint to redirect.
func (tun *SshTunnel) SetLocalEndpoint(endpoint *Endpoint) {
	tun.local = endpoint
}

// SetRemoteEndpoint sets the remote endpoint to redirect.
func (tun *SshTunnel) SetRemoteEndpoint(endpoint *Endpoint) {
	tun.remote = endpoint
}

// SetTimeout sets the connection timeouts (defaults to 15 seconds).
func (tun *SshTunnel) SetTimeout(timeout time.Duration) {
	tun.timeout = timeout
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
// Note on SSH authentication: in case the tunnel's authType is set to AuthTypeAuto the following will happen:
// The default key files will be used, if that doesn't succeed it will try to use the SSH server.
// If that fails the whole authentication fails.
// That means if you want to use password or encrypted key file authentication, you have to specify that explicitly.
func (tun *SshTunnel) Start(ctx context.Context) error {
	tun.mutex.Lock()
	if tun.started {
		tun.mutex.Unlock()
		return fmt.Errorf("already started")
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
		User: tun.user,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: tun.timeout,
	}

	authMethod, err := tun.getSSHAuthMethod()
	if err != nil {
		return nil, err
	}

	config.Auth = []ssh.AuthMethod{authMethod}

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
		sshClient, err := ssh.Dial(tun.server.Type(), tun.server.String(), tun.sshConfig)
		if err != nil {
			return fmt.Errorf("ssh dial %s to %s failed: %w", tun.server.Type(), tun.server.String(), err)
		}
		tun.sshClient = sshClient
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
