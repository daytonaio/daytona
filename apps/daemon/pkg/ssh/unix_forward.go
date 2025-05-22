// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package ssh

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"

	log "github.com/sirupsen/logrus"
)

// streamLocalForwardPayload describes the extra data sent in a
// streamlocal-forward@openssh.com containing the socket path to bind to.
type streamLocalForwardPayload struct {
	SocketPath string
}

// forwardedStreamLocalPayload describes the data sent as the payload in the new
// channel request when a Unix connection is accepted by the listener.
type forwardedStreamLocalPayload struct {
	SocketPath string
	Reserved   uint32
}

// forwardedUnixHandler is a clone of ssh.ForwardedTCPHandler that does
// streamlocal forwarding (aka. unix forwarding) instead of TCP forwarding.
type forwardedUnixHandler struct {
	sync.Mutex
	forwards map[forwardKey]net.Listener
}

type forwardKey struct {
	sessionID string
	addr      string
}

func newForwardedUnixHandler() *forwardedUnixHandler {
	return &forwardedUnixHandler{
		forwards: make(map[forwardKey]net.Listener),
	}
}

func (h *forwardedUnixHandler) HandleSSHRequest(ctx ssh.Context, _ *ssh.Server, req *gossh.Request) (bool, []byte) {
	log.Debug(ctx, "handling SSH unix forward")
	conn, ok := ctx.Value(ssh.ContextKeyConn).(*gossh.ServerConn)
	if !ok {
		log.Warn(ctx, "SSH unix forward request from client with no gossh connection")
		return false, nil
	}

	switch req.Type {
	case "streamlocal-forward@openssh.com":
		var reqPayload streamLocalForwardPayload
		err := gossh.Unmarshal(req.Payload, &reqPayload)
		if err != nil {
			log.Warn(ctx, "parse streamlocal-forward@openssh.com request (SSH unix forward) payload from client", err)
			return false, nil
		}

		addr := reqPayload.SocketPath
		log.Debug(ctx, "request begin SSH unix forward")

		key := forwardKey{
			sessionID: ctx.SessionID(),
			addr:      addr,
		}

		h.Lock()
		_, ok := h.forwards[key]
		h.Unlock()
		if ok {
			// In cases where `ExitOnForwardFailure=yes` is set, returning false
			// here will cause the connection to be closed. To avoid this, and
			// to match OpenSSH behavior, we silently ignore the second forward
			// request.
			log.Warn(ctx, "SSH unix forward request for socket path that is already being forwarded on this session, ignoring")
			return true, nil
		}

		// Create socket parent dir if not exists.
		parentDir := filepath.Dir(addr)
		err = os.MkdirAll(parentDir, 0o700)
		if err != nil {
			log.Error(err)
			return false, nil
		}

		// Remove existing socket if it exists. We do not use os.Remove() here
		// so that directories are kept. Note that it's possible that we will
		// overwrite a regular file here. Both of these behaviors match OpenSSH,
		// however, which is why we unlink.
		err = unlink(addr)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			log.Warn(ctx, "remove existing socket for SSH unix forward request", err)
			return false, nil
		}

		lc := &net.ListenConfig{}
		ln, err := lc.Listen(ctx, "unix", addr)
		if err != nil {
			log.Warn(ctx, "listen on Unix socket for SSH unix forward request", err)
			return false, nil
		}
		log.Debug(ctx, "SSH unix forward listening on socket")

		// The listener needs to successfully start before it can be added to
		// the map, so we don't have to worry about checking for an existing
		// listener.
		//
		// This is also what the upstream TCP version of this code does.
		h.Lock()
		h.forwards[key] = ln
		h.Unlock()
		log.Debug(ctx, "SSH unix forward added to cache")

		ctx, cancel := context.WithCancel(ctx)
		go func() {
			<-ctx.Done()
			_ = ln.Close()
		}()
		go func() {
			defer cancel()

			for {
				c, err := ln.Accept()
				if err != nil {
					if !errors.Is(err, net.ErrClosed) {
						log.Warn(ctx, "accept on local Unix socket for SSH unix forward request", err)
					}
					// closed below
					log.Debug(ctx, "SSH unix forward listener closed")
					break
				}
				log.Debug(ctx, "accepted SSH unix forward connection")
				payload := gossh.Marshal(&forwardedStreamLocalPayload{
					SocketPath: addr,
				})

				go func() {
					ch, reqs, err := conn.OpenChannel("forwarded-streamlocal@openssh.com", payload)
					if err != nil {
						log.Warn(ctx, "open SSH unix forward channel to client", err)
						_ = c.Close()
						return
					}
					go gossh.DiscardRequests(reqs)
					Bicopy(ctx, ch, c)
				}()
			}

			h.Lock()
			if ln2, ok := h.forwards[key]; ok && ln2 == ln {
				delete(h.forwards, key)
			}
			h.Unlock()
			log.Debug(ctx, "SSH unix forward listener removed from cache")
			_ = ln.Close()
		}()

		return true, nil

	case "cancel-streamlocal-forward@openssh.com":
		var reqPayload streamLocalForwardPayload
		err := gossh.Unmarshal(req.Payload, &reqPayload)
		if err != nil {
			log.Warn(ctx, "parse cancel-streamlocal-forward@openssh.com (SSH unix forward) request payload from client", err)
			return false, nil
		}
		log.Debug(ctx, "request to cancel SSH unix forward", reqPayload.SocketPath)

		key := forwardKey{
			sessionID: ctx.SessionID(),
			addr:      reqPayload.SocketPath,
		}

		h.Lock()
		ln, ok := h.forwards[key]
		delete(h.forwards, key)
		h.Unlock()
		if !ok {
			log.Warn(ctx, "SSH unix forward not found in cache")
			return true, nil
		}
		_ = ln.Close()
		return true, nil

	default:
		return false, nil
	}
}

// directStreamLocalPayload describes the extra data sent in a
// direct-streamlocal@openssh.com channel request containing the socket path.
type directStreamLocalPayload struct {
	SocketPath string

	Reserved1 string
	Reserved2 uint32
}

func directStreamLocalHandler(_ *ssh.Server, _ *gossh.ServerConn, newChan gossh.NewChannel, ctx ssh.Context) {
	var reqPayload directStreamLocalPayload
	err := gossh.Unmarshal(newChan.ExtraData(), &reqPayload)
	if err != nil {
		_ = newChan.Reject(gossh.ConnectionFailed, "could not parse direct-streamlocal@openssh.com channel payload")
		return
	}

	var dialer net.Dialer
	dconn, err := dialer.DialContext(ctx, "unix", reqPayload.SocketPath)
	if err != nil {
		_ = newChan.Reject(gossh.ConnectionFailed, fmt.Sprintf("dial unix socket %q: %+v", reqPayload.SocketPath, err.Error()))
		return
	}

	ch, reqs, err := newChan.Accept()
	if err != nil {
		_ = dconn.Close()
		return
	}
	go gossh.DiscardRequests(reqs)

	Bicopy(ctx, ch, dconn)
}

// unlink removes files and unlike os.Remove, directories are kept.
func unlink(path string) error {
	// Ignore EINTR like os.Remove, see ignoringEINTR in os/file_posix.go
	// for more details.
	for {
		err := syscall.Unlink(path)
		if !errors.Is(err, syscall.EINTR) {
			return err
		}
	}
}

// Bicopy copies all of the data between the two connections and will close them
// after one or both of them are done writing. If the context is canceled, both
// of the connections will be closed.
func Bicopy(ctx context.Context, c1, c2 io.ReadWriteCloser) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	defer func() {
		_ = c1.Close()
		_ = c2.Close()
	}()

	var wg sync.WaitGroup
	copyFunc := func(dst io.WriteCloser, src io.Reader) {
		defer func() {
			wg.Done()
			// If one side of the copy fails, ensure the other one exits as
			// well.
			cancel()
		}()
		_, _ = io.Copy(dst, src)
	}

	wg.Add(2)
	go copyFunc(c1, c2)
	go copyFunc(c2, c1)

	// Convert waitgroup to a channel so we can also wait on the context.
	done := make(chan struct{})
	go func() {
		defer close(done)
		wg.Wait()
	}()

	select {
	case <-ctx.Done():
	case <-done:
	}
}
