// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package ssh

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"

	"github.com/daytonaio/daemon/pkg/common"
	"github.com/daytonaio/daemon/pkg/ssh/config"
	"github.com/gliderlabs/ssh"
	"github.com/pkg/sftp"
	"golang.org/x/sys/unix"
)

type Server struct {
	logger         *slog.Logger
	workDir        string
	defaultWorkDir string
}

func NewServer(logger *slog.Logger, workDir, defaultWorkDir string) *Server {
	return &Server{
		logger:         logger.With(slog.String("component", "ssh_server")),
		workDir:        workDir,
		defaultWorkDir: defaultWorkDir,
	}
}

func (s *Server) Start() error {
	forwardedTCPHandler := &ssh.ForwardedTCPHandler{}
	unixForwardHandler := newForwardedUnixHandler()

	sshServer := ssh.Server{
		Addr: fmt.Sprintf(":%d", config.SSH_PORT),
		PublicKeyHandler: func(ctx ssh.Context, key ssh.PublicKey) bool {
			// Allow all public key authentication attempts
			s.logger.Debug("Public key authentication accepted", "user", ctx.User())
			return true
		},
		PasswordHandler: func(ctx ssh.Context, password string) bool {
			s.logger.Debug("Password authentication attempt", "user", ctx.User())
			if len(password) > 0 {
				s.logger.Debug("Received password", "length", len(password))
			} else {
				s.logger.Debug("Received empty password")
			}
			// Only allow authentication with the hardcoded password 'sandbox-ssh'
			authenticated := password == "sandbox-ssh"
			if authenticated {
				s.logger.Debug("Password authentication succeeded", "user", ctx.User())
			} else {
				s.logger.Debug("Password authentication failed (wrong password)", "user", ctx.User())
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
				s.logger.Error("Subsystem not supported", "subsystem", ss)
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

	s.logger.Info("Starting ssh server", "port", config.SSH_PORT)
	return sshServer.ListenAndServe()
}

func (s *Server) handlePty(session ssh.Session, ptyReq ssh.Pty, winCh <-chan ssh.Window) {
	dir := s.workDir

	if _, err := os.Stat(s.workDir); os.IsNotExist(err) {
		dir = s.defaultWorkDir
	}

	env := []string{}

	if ssh.AgentRequested(session) {
		l, err := ssh.NewAgentListener()
		if err != nil {
			s.logger.Error("Failed to start agent listener", "error", err)
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
		s.logger.Debug("Failed to spawn tty", "error", err)
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
			s.logger.Error("Failed to start agent listener", "error", err)
			return
		}
		defer l.Close()
		go ssh.ForwardAgentConnections(l, session)
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", "SSH_AUTH_SOCK", l.Addr().String()))
	}

	cmd.Dir = s.workDir
	if _, err := os.Stat(s.workDir); os.IsNotExist(err) {
		cmd.Dir = s.defaultWorkDir
	}

	cmd.Stdout = session
	cmd.Stderr = session.Stderr()
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		s.logger.Error("Unable to setup stdin for session", "error", err)
		return
	}
	go func() {
		_, err := io.Copy(stdinPipe, session)
		if err != nil {
			s.logger.Error("Unable to read from session", "error", err)
			return
		}
		_ = stdinPipe.Close()
	}()

	err = cmd.Start()
	if err != nil {
		s.logger.Error("Unable to start command", "error", err)
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
				s.logger.Warn("Unable to send signal to process", "error", err)
			}
		}
	}()
	err = cmd.Wait()

	if err != nil {
		s.logger.Info("Command exited", "command", session.RawCommand(), "error", err)
		session.Exit(127)
		return
	}

	err = session.Exit(0)
	if err != nil {
		s.logger.Warn("Unable to exit session", "error", err)
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
		s.logger.Error("sftp server init error", "error", err)
		return
	}
	if err := server.Serve(); err == io.EOF {
		server.Close()
	} else if err != nil {
		s.logger.Error("sftp server completed with error", "error", err)
	}
}
