// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

// SSH Backdoor - Minimal SSH server for golden image maintenance
// This binary provides SSH access to VMs independent of the main daemon
// Password: sandbox-ssh (hardcoded)
// Port: 2222

package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
	"github.com/pkg/sftp"
	log "github.com/sirupsen/logrus"
)

const (
	SSH_PORT = 2222
	PASSWORD = "sandbox-ssh"
)

type TTYSize struct {
	Height int
	Width  int
}

func main() {
	log.SetLevel(log.DebugLevel)
	log.Printf("Starting SSH backdoor on port %d...", SSH_PORT)
	log.Printf("Password: %s", PASSWORD)

	server := &Server{
		WorkDir:        "/home/daytona",
		DefaultWorkDir: "/",
	}

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start SSH server: %v", err)
	}
}

type Server struct {
	WorkDir        string
	DefaultWorkDir string
}

func (s *Server) Start() error {
	sshServer := ssh.Server{
		Addr: fmt.Sprintf(":%d", SSH_PORT),
		PublicKeyHandler: func(ctx ssh.Context, key ssh.PublicKey) bool {
			log.Debugf("Public key authentication accepted for user: %s", ctx.User())
			return true
		},
		PasswordHandler: func(ctx ssh.Context, password string) bool {
			log.Debugf("Password authentication attempt for user: %s", ctx.User())
			authenticated := password == PASSWORD
			if authenticated {
				log.Debugf("Password authentication succeeded for user: %s", ctx.User())
			} else {
				log.Debugf("Password authentication failed for user: %s", ctx.User())
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
				log.Errorf("Subsystem %s not supported", ss)
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
		SubsystemHandlers: map[string]ssh.SubsystemHandler{
			"sftp": s.sftpHandler,
		},
		LocalPortForwardingCallback: ssh.LocalPortForwardingCallback(func(ctx ssh.Context, dhost string, dport uint32) bool {
			return true
		}),
	}

	return sshServer.ListenAndServe()
}

func (s *Server) handlePty(session ssh.Session, ptyReq ssh.Pty, winCh <-chan ssh.Window) {
	dir := s.WorkDir
	if _, err := os.Stat(s.WorkDir); os.IsNotExist(err) {
		dir = s.DefaultWorkDir
	}

	shell := getShell()
	cmd := exec.Command(shell)
	cmd.Dir = dir
	cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", ptyReq.Term))
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, fmt.Sprintf("SHELL=%s", shell))

	if ssh.AgentRequested(session) {
		l, err := ssh.NewAgentListener()
		if err != nil {
			log.Errorf("Failed to start agent listener: %v", err)
			return
		}
		defer l.Close()
		go ssh.ForwardAgentConnections(l, session)
		cmd.Env = append(cmd.Env, fmt.Sprintf("SSH_AUTH_SOCK=%s", l.Addr().String()))
	}

	f, err := pty.Start(cmd)
	if err != nil {
		log.Errorf("Failed to start pty: %v", err)
		return
	}
	defer f.Close()

	go func() {
		for win := range winCh {
			syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
				uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(win.Height), uint16(win.Width), 0, 0})))
		}
	}()

	go func() {
		io.Copy(f, session)
	}()

	io.Copy(session, f)
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
		cmd.Env = append(cmd.Env, fmt.Sprintf("SSH_AUTH_SOCK=%s", l.Addr().String()))
	}

	cmd.Dir = s.WorkDir
	if _, err := os.Stat(s.WorkDir); os.IsNotExist(err) {
		cmd.Dir = s.DefaultWorkDir
	}

	cmd.Stdout = session
	cmd.Stderr = session.Stderr()
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		log.Errorf("Unable to setup stdin: %v", err)
		return
	}

	go func() {
		io.Copy(stdinPipe, session)
		stdinPipe.Close()
	}()

	if err := cmd.Run(); err != nil {
		log.Debugf("Command exited with error: %v", err)
		session.Exit(127)
		return
	}

	session.Exit(0)
}

func (s *Server) sftpHandler(session ssh.Session) {
	server, err := sftp.NewServer(session)
	if err != nil {
		log.Errorf("sftp server init error: %s", err)
		return
	}
	if err := server.Serve(); err == io.EOF {
		server.Close()
	} else if err != nil {
		log.Errorf("sftp server error: %s", err)
	}
}

func getShell() string {
	out, err := exec.Command("sh", "-c", "grep '^[^#]' /etc/shells").Output()
	if err != nil {
		return "sh"
	}

	shells := string(out)
	if strings.Contains(shells, "/usr/bin/zsh") {
		return "/usr/bin/zsh"
	}
	if strings.Contains(shells, "/bin/zsh") {
		return "/bin/zsh"
	}
	if strings.Contains(shells, "/usr/bin/bash") {
		return "/usr/bin/bash"
	}
	if strings.Contains(shells, "/bin/bash") {
		return "/bin/bash"
	}
	return "/bin/sh"
}
