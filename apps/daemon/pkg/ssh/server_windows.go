//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package ssh

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"

	"github.com/daytonaio/daemon/pkg/childreap"
	"github.com/daytonaio/daemon/pkg/common"
	"github.com/daytonaio/daemon/pkg/ssh/config"
	"github.com/gliderlabs/ssh"
	"github.com/pkg/sftp"
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

	sshServer := ssh.Server{
		Addr: fmt.Sprintf(":%d", config.SSH_PORT),
		PublicKeyHandler: func(ctx ssh.Context, key ssh.PublicKey) bool {
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
			"session":      ssh.DefaultSessionHandler,
			"direct-tcpip": ssh.DirectTCPIPHandler,
		},
		RequestHandlers: map[string]ssh.RequestHandler{
			"tcpip-forward":        forwardedTCPHandler.HandleSSHRequest,
			"cancel-tcpip-forward": forwardedTCPHandler.HandleSSHRequest,
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

	s.logger.Info("Starting SSH server", "port", config.SSH_PORT)
	return sshServer.ListenAndServe()
}

func (s *Server) handlePty(session ssh.Session, ptyReq ssh.Pty, winCh <-chan ssh.Window) {
	dir := s.workDir

	if _, err := os.Stat(s.workDir); os.IsNotExist(err) {
		dir = s.defaultWorkDir
	}

	sizeCh := make(chan common.TTYSize)
	go func() {
		defer close(sizeCh)
		for win := range winCh {
			sizeCh <- common.TTYSize{Width: win.Width, Height: win.Height}
		}
	}()

	err := common.SpawnTTY(common.SpawnTTYOptions{
		Ctx:      session.Context(),
		Dir:      dir,
		StdIn:    session,
		StdOut:   session,
		InitCols: ptyReq.Window.Width,
		InitRows: ptyReq.Window.Height,
		SizeCh:   sizeCh,
	})
	if err != nil {
		s.logger.Debug("Failed to spawn PTY", "error", err)
		return
	}
}

func (s *Server) handleNonPty(session ssh.Session) {
	shell := common.GetShell()

	var cmd *exec.Cmd
	if len(session.Command()) > 0 {
		rawCmd := session.RawCommand()

		parsedCommand, envVars := common.ParseShellWrapper(rawCmd)
		if parsedCommand != rawCmd {
			s.logger.Debug("Parsed shell wrapper", "raw", rawCmd, "parsed", parsedCommand, "env", envVars)
		}

		finalCommand := common.BuildWindowsCommandForShell(parsedCommand, envVars, common.IsPowerShell(shell))

		cmd = common.ShellCommand(shell, finalCommand)
	} else {
		cmd = exec.Command(shell)
	}
	cmd.Env = append(cmd.Env, os.Environ()...)
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
		if _, err := io.Copy(stdinPipe, session); err != nil {
			s.logger.Error("Unable to read from session", "error", err)
			return
		}
		_ = stdinPipe.Close()
	}()

	if err := cmd.Start(); err != nil {
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
			s.handleSignal(cmd, sig)
		}
	}()

	exitCode, waitErr := childreap.Wait(cmd)

	if waitErr != nil || exitCode != 0 {
		s.logger.Info("Command exited", "command", session.RawCommand(), "exitCode", exitCode, "error", waitErr)
		// childreap.Wait returns -1 when no exit status could be recovered.
		// The SSH protocol carries exit status as uint32, so a negative
		// value gets serialized as 4294967295 — confusing to clients.
		// Normalize any negative to a generic non-zero exit.
		exitStatus := exitCode
		if exitStatus < 0 {
			exitStatus = 1
		}
		session.Exit(exitStatus)
		return
	}

	if err := session.Exit(0); err != nil {
		s.logger.Warn("Unable to exit session", "error", err)
	}
}

func (s *Server) handleSignal(cmd *exec.Cmd, sig ssh.Signal) {
	if cmd.Process == nil {
		return
	}

	switch sig {
	case ssh.SIGKILL, ssh.SIGTERM, ssh.SIGINT, ssh.SIGQUIT:
		if err := cmd.Process.Kill(); err != nil {
			s.logger.Warn("Unable to kill process", "error", err)
		}
	default:
		s.logger.Debug("Signal received, killing process on Windows", "signal", sig)
		if err := cmd.Process.Kill(); err != nil {
			s.logger.Warn("Unable to kill process for signal", "signal", sig, "error", err)
		}
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
