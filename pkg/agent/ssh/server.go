// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ssh

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/creack/pty"
	"github.com/daytonaio/daytona/pkg/agent/ssh/config"
	"github.com/gliderlabs/ssh"
	"github.com/pkg/sftp"

	log "github.com/sirupsen/logrus"
)

type Server struct {
	ProjectDir        string
	DefaultProjectDir string
}

func (s *Server) Start() error {
	forwardedTCPHandler := &ssh.ForwardedTCPHandler{}
	unixForwardHandler := newForwardedUnixHandler()

	sshServer := ssh.Server{
		Addr: fmt.Sprintf(":%d", config.SSH_PORT),
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
	shell := s.getShell()
	cmd := exec.Command(shell)

	cmd.Dir = s.ProjectDir

	if _, err := os.Stat(s.ProjectDir); os.IsNotExist(err) {
		cmd.Dir = s.DefaultProjectDir
	}

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

	cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", ptyReq.Term))
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, fmt.Sprintf("SHELL=%s", shell))
	f, err := pty.Start(cmd)
	if err != nil {
		log.Errorf("Unable to start command: %v", err)
		return
	}

	go func() {
		for win := range winCh {
			SetPtySize(f, win)
		}
	}()
	go func() {
		io.Copy(f, session) // stdin
	}()
	io.Copy(session, f) // stdout
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
			signal := OsSignalFrom(sig)
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

func (s *Server) getShell() string {
	out, err := exec.Command("sh", "-c", "grep '^[^#]' /etc/shells").Output()
	if err != nil {
		return "sh"
	}

	if strings.Contains(string(out), "/usr/bin/zsh") {
		return "/usr/bin/zsh"
	}

	if strings.Contains(string(out), "/bin/zsh") {
		return "/bin/zsh"
	}

	if strings.Contains(string(out), "/usr/bin/bash") {
		return "/usr/bin/bash"
	}

	if strings.Contains(string(out), "/bin/bash") {
		return "/bin/bash"
	}

	shellEnv, shellSet := os.LookupEnv("SHELL")

	if shellSet {
		return shellEnv
	}

	return "sh"
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
