// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package ssh

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/daytonaio/daemon/pkg/common"
	"github.com/daytonaio/daemon/pkg/ssh/config"
	"github.com/gliderlabs/ssh"
	"github.com/pkg/sftp"
	"golang.org/x/sys/unix"

	log "github.com/sirupsen/logrus"
)

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type Server struct {
	ProjectDir        string
	DefaultProjectDir string
}

func (s *Server) Start() error {
	forwardedTCPHandler := &ssh.ForwardedTCPHandler{}
	unixForwardHandler := newForwardedUnixHandler()

	sshServer := ssh.Server{
		Addr: fmt.Sprintf(":%d", config.SSH_PORT),
		PublicKeyHandler: func(ctx ssh.Context, key ssh.PublicKey) bool {
			// Allow all public key authentication attempts
			log.Debugf("Public key authentication accepted for user: %s", ctx.User())
			return true
		},
		PasswordHandler: func(ctx ssh.Context, password string) bool {
			log.Debugf("Password authentication attempt for user: %s", ctx.User())
			if len(password) > 0 {
				log.Debugf("Received password length: %d, starts with: %s", len(password), password[:min(len(password), 3)])
			} else {
				log.Debugf("Received empty password")
			}
			// Only allow authentication with the hardcoded password 'sandbox-ssh'
			authenticated := password == "sandbox-ssh"
			if authenticated {
				log.Debugf("Password authentication succeeded for user: %s", ctx.User())
			} else {
				log.Debugf("Password authentication failed for user: %s (wrong password)", ctx.User())
			}
			return authenticated
		},
		Handler: func(session ssh.Session) {
			switch ss := session.Subsystem(); ss {
			case "":
			case "sftp":
				s.sftpHandler(session)
				return
			default:
				log.Errorf("Subsystem %s not supported\n", ss)
				session.Exit(1)
				return
			}

			ptyReq, winCh, isPty := session.Pty()
			if session.RawCommand() == "" && isPty {
				s.handlePty(session, ptyReq, winCh)
			} else {
				s.handleNonPty(session)
			}
		},
		ChannelHandlers: map[string]ssh.ChannelHandler{
			"session":                        ssh.DefaultSessionHandler,
			"direct-tcpip":                   ssh.DirectTCPIPHandler,
			"direct-streamlocal@openssh.com": directStreamLocalHandler,
		},
		RequestHandlers: map[string]ssh.RequestHandler{
			"tcpip-forward":                          forwardedTCPHandler.HandleSSHRequest,
			"cancel-tcpip-forward":                   forwardedTCPHandler.HandleSSHRequest,
			"streamlocal-forward@openssh.com":        unixForwardHandler.HandleSSHRequest,
			"cancel-streamlocal-forward@openssh.com": unixForwardHandler.HandleSSHRequest,
		},
		SubsystemHandlers: map[string]ssh.SubsystemHandler{
			"sftp": s.sftpHandler,
		},
		LocalPortForwardingCallback: ssh.LocalPortForwardingCallback(func(ctx ssh.Context, dhost string, dport uint32) bool {
			return true
		}),
		ReversePortForwardingCallback: ssh.ReversePortForwardingCallback(func(ctx ssh.Context, host string, port uint32) bool {
			return true
		}),
		SessionRequestCallback: func(sess ssh.Session, requestType string) bool {
			return true
		},
	}

	log.Printf("Starting ssh server on port %d...\n", config.SSH_PORT)
	return sshServer.ListenAndServe()
}

func (s *Server) handlePty(session ssh.Session, ptyReq ssh.Pty, winCh <-chan ssh.Window) {
	dir := s.ProjectDir

	if _, err := os.Stat(s.ProjectDir); os.IsNotExist(err) {
		dir = s.DefaultProjectDir
	}

	env := []string{}

	if ssh.AgentRequested(session) {
		l, err := ssh.NewAgentListener()
		if err != nil {
			log.Errorf("Failed to start agent listener: %v", err)
			return
		}
		defer l.Close()
		go ssh.ForwardAgentConnections(l, session)
		env = append(env, fmt.Sprintf("%s=%s", "SSH_AUTH_SOCK", l.Addr().String()))
	}

	sizeCh := make(chan common.TTYSize)

	go func() {
		for win := range winCh {
			sizeCh <- common.TTYSize{
				Height: win.Height,
				Width:  win.Width,
			}
		}
	}()

	err := common.SpawnTTY(common.SpawnTTYOptions{
		Dir:    dir,
		StdIn:  session,
		StdOut: session,
		Term:   ptyReq.Term,
		Env:    env,
		SizeCh: sizeCh,
	})

	if err != nil {
		// Debug log here because this gets called on each ssh "exit"
		// TODO: Find a better way to handle this
		log.Debugf("Failed to spawn tty: %v", err)
		return
	}
}

func (s *Server) handleNonPty(session ssh.Session) {
	args := []string{}
	if len(session.Command()) > 0 {
		args = append([]string{"-c"}, session.RawCommand())
	}

	cmd := exec.Command("/bin/sh", args...)

	cmd.Env = append(cmd.Env, os.Environ()...)

	if ssh.AgentRequested(session) {
		l, err := ssh.NewAgentListener()
		if err != nil {
			log.Errorf("Failed to start agent listener: %v", err)
			return
		}
		defer l.Close()
		go ssh.ForwardAgentConnections(l, session)
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", "SSH_AUTH_SOCK", l.Addr().String()))
	}

	cmd.Dir = s.ProjectDir
	if _, err := os.Stat(s.ProjectDir); os.IsNotExist(err) {
		cmd.Dir = s.DefaultProjectDir
	}

	cmd.Stdout = session
	cmd.Stderr = session.Stderr()
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		log.Errorf("Unable to setup stdin for session: %v", err)
		return
	}
	go func() {
		_, err := io.Copy(stdinPipe, session)
		if err != nil {
			log.Errorf("Unable to read from session: %v", err)
			return
		}
		_ = stdinPipe.Close()
	}()

	err = cmd.Start()
	if err != nil {
		log.Errorf("Unable to start command: %v", err)
		return
	}
	sigs := make(chan ssh.Signal, 1)
	session.Signals(sigs)
	defer func() {
		session.Signals(nil)
		close(sigs)
	}()
	go func() {
		for sig := range sigs {
			signal := s.osSignalFrom(sig)
			err := cmd.Process.Signal(signal)
			if err != nil {
				log.Warnf("Unable to send signal to process: %v", err)
			}
		}
	}()
	err = cmd.Wait()

	if err != nil {
		log.Println(session.RawCommand(), " ", err)
		session.Exit(127)
		return
	}

	err = session.Exit(0)
	if err != nil {
		log.Warnf("Unable to exit session: %v", err)
	}
}

func (s *Server) osSignalFrom(sig ssh.Signal) os.Signal {
	switch sig {
	case ssh.SIGABRT:
		return unix.SIGABRT
	case ssh.SIGALRM:
		return unix.SIGALRM
	case ssh.SIGFPE:
		return unix.SIGFPE
	case ssh.SIGHUP:
		return unix.SIGHUP
	case ssh.SIGILL:
		return unix.SIGILL
	case ssh.SIGINT:
		return unix.SIGINT
	case ssh.SIGKILL:
		return unix.SIGKILL
	case ssh.SIGPIPE:
		return unix.SIGPIPE
	case ssh.SIGQUIT:
		return unix.SIGQUIT
	case ssh.SIGSEGV:
		return unix.SIGSEGV
	case ssh.SIGTERM:
		return unix.SIGTERM
	case ssh.SIGUSR1:
		return unix.SIGUSR1
	case ssh.SIGUSR2:
		return unix.SIGUSR2

	// Unhandled, use sane fallback.
	default:
		return unix.SIGKILL
	}
}

func (s *Server) sftpHandler(session ssh.Session) {
	debugStream := io.Discard
	serverOptions := []sftp.ServerOption{
		sftp.WithDebug(debugStream),
	}
	server, err := sftp.NewServer(
		session,
		serverOptions...,
	)
	if err != nil {
		log.Errorf("sftp server init error: %s\n", err)
		return
	}
	if err := server.Serve(); err == io.EOF {
		server.Close()
	} else if err != nil {
		log.Errorf("sftp server completed with error: %s\n", err)
	}
}
