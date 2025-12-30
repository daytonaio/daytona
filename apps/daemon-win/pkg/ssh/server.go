// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package ssh

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/daytonaio/daemon-win/pkg/common"
	"github.com/daytonaio/daemon-win/pkg/ssh/config"
	"github.com/gliderlabs/ssh"
	"github.com/pkg/sftp"

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
	WorkDir        string
	DefaultWorkDir string
}

func (s *Server) Start() error {
	forwardedTCPHandler := &ssh.ForwardedTCPHandler{}

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

	log.Printf("Starting SSH server on port %d...\n", config.SSH_PORT)
	return sshServer.ListenAndServe()
}

func (s *Server) handlePty(session ssh.Session, ptyReq ssh.Pty, winCh <-chan ssh.Window) {
	dir := s.WorkDir

	if _, err := os.Stat(s.WorkDir); os.IsNotExist(err) {
		dir = s.DefaultWorkDir
	}

	// Use ConPTY for Windows PTY support
	err := SpawnConPTY(SpawnConPTYOptions{
		Dir:    dir,
		StdIn:  session,
		StdOut: session,
		Cols:   uint16(ptyReq.Window.Width),
		Rows:   uint16(ptyReq.Window.Height),
		WinCh:  winCh,
	})

	if err != nil {
		log.Debugf("Failed to spawn ConPTY: %v", err)
		return
	}
}

func (s *Server) handleNonPty(session ssh.Session) {
	shell := common.GetShell()
	shellArgs := common.GetShellArgs(shell)

	var args []string
	if len(session.Command()) > 0 {
		// Parse the command for Windows compatibility
		rawCmd := session.RawCommand()

		// Check if this is a Linux-style shell wrapper (from SDKs)
		parsedCommand, envVars := common.ParseShellWrapper(rawCmd)
		if parsedCommand != rawCmd {
			log.Debugf("Parsed shell wrapper: %q -> %q (env: %v)", rawCmd, parsedCommand, envVars)
		}

		// Build Windows command with env vars if any
		finalCommand := common.BuildWindowsCommand(parsedCommand, envVars)

		args = append(shellArgs, finalCommand)
	}

	cmd := exec.Command(shell, args...)
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Dir = s.WorkDir

	if _, err := os.Stat(s.WorkDir); os.IsNotExist(err) {
		cmd.Dir = s.DefaultWorkDir
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
			s.handleSignal(cmd, sig)
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

// handleSignal handles SSH signals on Windows
func (s *Server) handleSignal(cmd *exec.Cmd, sig ssh.Signal) {
	if cmd.Process == nil {
		return
	}

	// On Windows, we can only reliably kill processes
	switch sig {
	case ssh.SIGKILL, ssh.SIGTERM, ssh.SIGINT, ssh.SIGQUIT:
		err := cmd.Process.Kill()
		if err != nil {
			log.Warnf("Unable to kill process: %v", err)
		}
	default:
		// For other signals, attempt to kill as Windows doesn't support Unix signals
		log.Debugf("Signal %s received, killing process on Windows", sig)
		err := cmd.Process.Kill()
		if err != nil {
			log.Warnf("Unable to kill process for signal %s: %v", sig, err)
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
		log.Errorf("sftp server init error: %s\n", err)
		return
	}
	if err := server.Serve(); err == io.EOF {
		server.Close()
	} else if err != nil {
		log.Errorf("sftp server completed with error: %s\n", err)
	}
}

// isPowerShell checks if the shell path refers to PowerShell
func isPowerShell(shell string) bool {
	shell = strings.ToLower(shell)
	return strings.Contains(shell, "pwsh") || strings.Contains(shell, "powershell")
}
