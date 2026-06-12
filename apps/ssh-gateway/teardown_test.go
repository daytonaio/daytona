// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package main

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
	// connClosed is closed when the client connection dies — the signal the
	// production watcher goroutine selects on to hard-close the upstream channel.
	connClosed chan struct{}
	// closeConns tears down both sides of the underlying TCP connection.
	closeConns func()
	// closeClient closes only the client-side TCP connection, simulating an
	// abrupt SSH client disconnect (SIGKILL / network drop). This closes
	// connClosed.
	closeClient func()
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
	serverConnCh := make(chan *ssh.ServerConn, 1)
	go func() {
		sconn, chans, reqs, err := ssh.NewServerConn(srvConn, serverCfg)
		if err != nil {
			return
		}
		serverConnCh <- sconn
		go ssh.DiscardRequests(reqs)
		newCh, ok := <-chans
		if !ok {
			return
		}
		ch, reqs, err := newCh.Accept()
		if err != nil {
			return
		}
		go ssh.DiscardRequests(reqs)
		serverChanCh <- ch
		for range chans {
		} // drain remaining channels
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

	// Mirror handleConnection: one Wait() goroutine per connection closes
	// connClosed for all per-channel watchers.
	connClosed := make(chan struct{})
	go func() {
		serverConn := <-serverConnCh
		_ = serverConn.Wait()
		close(connClosed)
	}()

	return &channelPair{
		server:     serverCh,
		client:     clientCh,
		connClosed: connClosed,
		closeConns: func() {
			cliConn.Close() //nolint:errcheck
			srvConn.Close() //nolint:errcheck
		},
		closeClient: func() {
			cliConn.Close() //nolint:errcheck
		},
	}
}

// TestClientDisconnectTeardownPropagatesUpstream is the regression test for the stale
// SSH keepalive bug (https://github.com/daytonaio/daytona/issues/4805).
//
// Before the fix, killing the SSH client (SIGKILL / network drop) left the reverse
// io.Copy (runner→client) blocked indefinitely: nothing closed runnerChannel, so the
// keepalive goroutine's context was never cancelled, lastActivityAt kept refreshing every
// 45 s, and auto-stop never triggered.
//
// The fix watches the client connection itself: when it dies, runnerChannel.Close()
// (MSG_CHANNEL_CLOSE) forces the runner to send its own MSG_CHANNEL_CLOSE in reply,
// unblocking the reverse io.Copy. A clean stdin EOF only half-closes (CloseWrite),
// preserving remaining output and exit-status.
func TestClientDisconnectTeardownPropagatesUpstream(t *testing.T) {
	t.Parallel()

	// clientPair simulates the gateway receiving a channel from the external SSH client.
	// clientPair.server == clientChannel in handleChannel (gateway's server-side view).
	// clientPair.closeClient() simulates SIGKILL: drops the TCP connection so
	// clientPair.connClosed fires, after which runnerChannel.Close() is called
	// in the production code.
	clientPair := newChannelPair(t)
	defer clientPair.closeConns()

	// runnerPair simulates the gateway opening a channel to the backend runner.
	// runnerPair.client == runnerChannel in handleChannel (gateway's client-side view).
	runnerPair := newChannelPair(t)
	defer runnerPair.closeConns()

	clientChannel := clientPair.server // gateway accepted this from the external client
	runnerChannel := runnerPair.client // gateway opened this to the runner

	// done is closed when the main blocking io.Copy (runner→client) returns,
	// proving that the keepalive context would be cancelled via defer cancel().
	done := make(chan struct{})

	// Mirror the goroutines from apps/ssh-gateway/main.go rather than calling handleChannel
	// directly: handleChannel requires a live SSH server listener and API client, making it
	// unsuitable for a unit test. The mechanism being tested (a dead client connection
	// hard-closes the runner channel, unblocking the reverse io.Copy) is fully captured
	// here without those dependencies.
	go func() {
		_, _ = io.Copy(runnerChannel, clientChannel)
		runnerChannel.CloseWrite() //nolint:errcheck
	}()
	sessionDone := make(chan struct{})
	defer close(sessionDone)
	go func() {
		select {
		case <-clientPair.connClosed:
			runnerChannel.Close() //nolint:errcheck
		case <-sessionDone:
		}
	}()

	// Main blocking copy: runner→client.
	// This is the call that previously blocked forever after a client SIGKILL.
	go func() {
		_, _ = io.Copy(clientChannel, runnerChannel)
		close(done)
	}()

	// Simulate abrupt SSH client disconnect (kill -9 / VS Code tab closed / network drop)
	// by closing the underlying TCP connection. This fires connClosed, which triggers
	// runnerChannel.Close().
	clientPair.closeClient()

	select {
	case <-done:
		// Both copies returned. In the real handleChannel, defer cancel() would now fire,
		// stopping the keepalive ticker. lastActivityAt goes stale → auto-stop triggers.
	case <-time.After(2 * time.Second):
		t.Fatal("reverse channel copy (runner→client) did not unblock after client disconnect; " +
			"the keepalive goroutine would run indefinitely, preventing auto-stop")
	}
}

// TestRunnerChannelClosedAfterClientDisconnect verifies that the runner-side channel
// receives a close signal after the external client disconnects. This ensures orphaned
// sandbox shells (VS Code remote server, JetBrains Gateway, plain bash) are cleaned up.
func TestRunnerChannelClosedAfterClientDisconnect(t *testing.T) {
	t.Parallel()

	clientPair := newChannelPair(t)
	defer clientPair.closeConns()

	runnerPair := newChannelPair(t)
	defer runnerPair.closeConns()

	clientChannel := clientPair.server
	runnerChannel := runnerPair.client

	go func() {
		_, _ = io.Copy(runnerChannel, clientChannel)
		runnerChannel.CloseWrite() //nolint:errcheck
	}()
	sessionDone := make(chan struct{})
	defer close(sessionDone)
	go func() {
		select {
		case <-clientPair.connClosed:
			runnerChannel.Close() //nolint:errcheck
		case <-sessionDone:
		}
	}()
	go func() {
		_, _ = io.Copy(clientChannel, runnerChannel)
	}()

	// Simulate abrupt SSH client disconnect (SIGKILL / network drop).
	clientPair.closeClient()

	// The runner-side server channel (runnerPair.server) should receive EOF/close,
	// which propagates to the runner's handleChannel → sandboxChannel.Close().
	readErr := make(chan error, 1)
	go func() {
		buf := make([]byte, 1)
		_, err := runnerPair.server.Read(buf)
		readErr <- err
	}()

	select {
	case err := <-readErr:
		if err == nil {
			t.Fatal("expected EOF/close on runner-side channel after client disconnect, got nil")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("runner-side channel was not closed after client disconnect; " +
			"orphaned sandbox processes would not receive a hangup signal")
	}
}

// TestStdinEOFPreservesHalfClose verifies that a clean stdin EOF from the client
// (e.g. `cat file | ssh host cmd`) only half-closes the runner channel: the remote
// command can still send output and exit-status afterwards. A hard Close() here would
// truncate them — the regression flagged in the PR review of the keepalive fix.
func TestStdinEOFPreservesHalfClose(t *testing.T) {
	t.Parallel()

	clientPair := newChannelPair(t)
	defer clientPair.closeConns()

	runnerPair := newChannelPair(t)
	defer runnerPair.closeConns()

	clientChannel := clientPair.server
	runnerChannel := runnerPair.client

	// Mirror the production goroutines from handleChannel.
	go func() {
		_, _ = io.Copy(runnerChannel, clientChannel)
		runnerChannel.CloseWrite() //nolint:errcheck
	}()
	sessionDone := make(chan struct{})
	defer close(sessionDone)
	go func() {
		select {
		case <-clientPair.connClosed:
			runnerChannel.Close() //nolint:errcheck
		case <-sessionDone:
		}
	}()
	go func() {
		_, _ = io.Copy(clientChannel, runnerChannel)
	}()

	// Clean stdin EOF: the client half-closes its channel while the SSH
	// connection stays alive.
	if err := clientPair.client.CloseWrite(); err != nil {
		t.Fatalf("client CloseWrite: %v", err)
	}

	// The runner side must observe EOF on its read...
	eofErr := make(chan error, 1)
	go func() {
		_, err := runnerPair.server.Read(make([]byte, 1))
		eofErr <- err
	}()
	select {
	case err := <-eofErr:
		if err != io.EOF {
			t.Fatalf("expected io.EOF on runner side after stdin EOF, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("runner side did not observe stdin EOF")
	}

	// ...and must still be able to send output, which the client receives.
	// With a hard Close() this write fails because the channel is torn down.
	want := "late output"
	if _, err := runnerPair.server.Write([]byte(want)); err != nil {
		t.Fatalf("runner write after stdin EOF failed (half-close broken): %v", err)
	}

	got := make([]byte, len(want))
	readDone := make(chan error, 1)
	go func() {
		_, err := io.ReadFull(clientPair.client, got)
		readDone <- err
	}()
	select {
	case err := <-readDone:
		if err != nil {
			t.Fatalf("client read after stdin EOF failed: %v", err)
		}
		if string(got) != want {
			t.Fatalf("client got %q, want %q", got, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("client did not receive output sent after stdin EOF; half-close broken")
	}
}
