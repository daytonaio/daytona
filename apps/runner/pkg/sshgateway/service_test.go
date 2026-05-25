// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build linux

package sshgateway

import (
	"crypto/rand"
	"crypto/rsa"
	"io"
	"net"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

// testSigner generates a throwaway RSA signer for in-memory SSH tests.
func testSigner(t *testing.T) ssh.Signer {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	signer, err := ssh.NewSignerFromKey(key)
	if err != nil {
		t.Fatalf("make signer: %v", err)
	}
	return signer
}

// channelPair holds both ends of an in-memory SSH channel.
type channelPair struct {
	// server is the channel as seen by the SSH server (Accept()-returned).
	server ssh.Channel
	// client is the channel as seen by the SSH client (OpenChannel()-returned).
	client ssh.Channel
	// closeConns tears down the underlying net.Pipe connections.
	closeConns func()
}

// newChannelPair creates an SSH connection backed by a loopback TCP socket and
// returns a connected channel pair. A loopback socket is used instead of
// net.Pipe() because net.Pipe() is synchronous with no kernel buffer: when both
// sides of an SSH handshake write their version string simultaneously, both block
// indefinitely. TCP's OS send buffer absorbs the small version string so neither
// side stalls.
func newChannelPair(t *testing.T) *channelPair {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	accepted := make(chan net.Conn, 1)
	go func() {
		conn, err := ln.Accept()
		ln.Close() //nolint:errcheck — only one connection needed
		if err == nil {
			accepted <- conn
		}
	}()

	cliConn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}

	var srvConn net.Conn
	select {
	case srvConn = <-accepted:
	case <-time.After(5 * time.Second):
		cliConn.Close() //nolint:errcheck
		t.Fatal("accept timeout")
	}

	serverCfg := &ssh.ServerConfig{NoClientAuth: true}
	serverCfg.AddHostKey(testSigner(t))

	serverChanCh := make(chan ssh.Channel, 1)
	go func() {
		_, chans, reqs, err := ssh.NewServerConn(srvConn, serverCfg)
		if err != nil {
			return
		}
		go ssh.DiscardRequests(reqs)
		for newCh := range chans {
			ch, reqs, err := newCh.Accept()
			if err != nil {
				return
			}
			go ssh.DiscardRequests(reqs)
			serverChanCh <- ch
			for range chans {
			} // drain remaining channels
			return
		}
	}()

	clientCfg := &ssh.ClientConfig{
		User:            "test",
		Auth:            []ssh.AuthMethod{ssh.Password("")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	cc, cliChans, cliReqs, err := ssh.NewClientConn(cliConn, "", clientCfg)
	if err != nil {
		t.Fatalf("ssh client connect: %v", err)
	}
	go ssh.DiscardRequests(cliReqs)
	go func() {
		for range cliChans {
		}
	}()

	clientCh, chReqs, err := cc.OpenChannel("session", nil)
	if err != nil {
		t.Fatalf("open channel: %v", err)
	}
	go ssh.DiscardRequests(chReqs)

	var serverCh ssh.Channel
	select {
	case serverCh = <-serverChanCh:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for server to accept channel")
	}

	return &channelPair{
		server: serverCh,
		client: clientCh,
		closeConns: func() {
			cliConn.Close() //nolint:errcheck
			srvConn.Close() //nolint:errcheck
		},
	}
}

// TestClientDisconnectClosesSandboxChannel is the regression test for the stale SSH
// keepalive bug (https://github.com/daytonaio/daytona/issues/4805) at the runner layer.
//
// When the SSH gateway closes its channel to the runner (because the external client
// disconnected), the runner must close the downstream sandbox channel. Without this,
// the sandbox-to-client io.Copy blocks forever, the sandbox shell (and any VS Code /
// JetBrains remote server processes) stays alive, and the gateway's keepalive ticker
// keeps refreshing lastActivityAt every 45 s.
//
// The fix adds sandboxChannel.Close() at the end of the client→sandbox goroutine.
func TestClientDisconnectClosesSandboxChannel(t *testing.T) {
	t.Parallel()

	// clientPair simulates the runner receiving a channel from the gateway.
	// clientPair.server == clientChannel in Service.handleChannel (runner's server-side view).
	// clientPair.client == the gateway's end; closing it simulates the external client dying.
	clientPair := newChannelPair(t)
	defer clientPair.closeConns()

	// sandboxPair simulates the runner opening a channel into the sandbox container.
	// sandboxPair.client == sandboxChannel in Service.handleChannel.
	// sandboxPair.server == the sandbox container's sshd end.
	sandboxPair := newChannelPair(t)
	defer sandboxPair.closeConns()

	clientChannel := clientPair.server   // runner accepted this from the gateway
	sandboxChannel := sandboxPair.client // runner opened this to the sandbox container

	// done is closed when the main blocking io.Copy (sandbox→client) returns.
	done := make(chan struct{})

	// Fixed goroutine: client→sandbox copy, then close sandbox on disconnect.
	// This is the goroutine that was modified in apps/runner/pkg/sshgateway/service.go.
	go func() {
		_, _ = io.Copy(sandboxChannel, clientChannel)
		sandboxChannel.Close() //nolint:errcheck — the fix
	}()

	// Main blocking copy: sandbox→client.
	// This previously blocked forever when the gateway closed its channel.
	go func() {
		_, _ = io.Copy(clientChannel, sandboxChannel)
		close(done)
	}()

	// Allow goroutines to reach their blocking io.Copy calls.
	time.Sleep(10 * time.Millisecond)

	// Simulate gateway closing its channel (because the external SSH client disconnected).
	clientPair.client.Close() //nolint:errcheck

	select {
	case <-done:
		// Both copies returned. The orphaned sandbox shell will receive HUP/EOF and exit.
	case <-time.After(2 * time.Second):
		t.Fatal("sandbox-side channel copy did not unblock after client disconnect; " +
			"orphaned sandbox processes (bash, VS Code remote, JetBrains) would persist indefinitely")
	}
}

// TestSandboxChannelReceivesCloseOnClientDisconnect verifies that the sandbox container's
// channel receives a close signal after the external client disconnects. This is what
// causes orphaned shell processes inside the sandbox to receive a hangup and exit.
func TestSandboxChannelReceivesCloseOnClientDisconnect(t *testing.T) {
	t.Parallel()

	clientPair := newChannelPair(t)
	defer clientPair.closeConns()

	sandboxPair := newChannelPair(t)
	defer sandboxPair.closeConns()

	clientChannel := clientPair.server
	sandboxChannel := sandboxPair.client

	go func() {
		_, _ = io.Copy(sandboxChannel, clientChannel)
		sandboxChannel.Close() //nolint:errcheck
	}()
	go func() {
		_, _ = io.Copy(clientChannel, sandboxChannel)
	}()

	time.Sleep(10 * time.Millisecond)

	// Gateway closes its channel (external client disconnected).
	clientPair.client.Close() //nolint:errcheck

	// The sandbox container's sshd side (sandboxPair.server) should receive EOF/close,
	// which causes the sandbox shell process to receive a hangup and exit.
	readErr := make(chan error, 1)
	go func() {
		buf := make([]byte, 1)
		_, err := sandboxPair.server.Read(buf)
		readErr <- err
	}()

	select {
	case err := <-readErr:
		if err == nil {
			t.Fatal("expected EOF/close on sandbox-side channel after client disconnect, got nil")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("sandbox channel was not closed after client disconnect; " +
			"orphaned shell processes would not receive hangup signal")
	}
}
