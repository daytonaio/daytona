// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ssh

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
	crypto_ssh "golang.org/x/crypto/ssh"
	"golang.org/x/sys/unix"

	log "github.com/sirupsen/logrus"
)

type Server struct {
	ProjectDir        string
	DefaultProjectDir string
}

func (s *Server) Start() error {
	forwardedTCPHandler := &ssh.ForwardedTCPHandler{}

	sshServer := ssh.Server{
		Addr: ":2222",
		Handler: func(session ssh.Session) {
			switch ss := session.Subsystem(); ss {
			case "":
			case "sftp":
				s.sftpHandler(session)
				return
			default:
				fmt.Fprintf(session, "Subsystem %s not supported\n", ss)
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
			"session": ssh.DefaultSessionHandler,
			"direct-tcpip": func(srv *ssh.Server, conn *crypto_ssh.ServerConn, newChan crypto_ssh.NewChannel, ctx ssh.Context) {
				ssh.DirectTCPIPHandler(srv, conn, newChan, ctx)
			},
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

	log.Println("starting ssh server on port 2222...")
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
			fmt.Println("#5", err.Error())
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
		panic(err)
	}

	go func() {
		for win := range winCh {
			syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
				uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(win.Height), uint16(win.Width), 0, 0})))
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

	fmt.Println("args: ", args)

	cmd := exec.Command("/bin/sh", args...)

	cmd.Env = append(cmd.Env, os.Environ()...)

	if ssh.AgentRequested(session) {
		l, err := ssh.NewAgentListener()
		if err != nil {
			fmt.Println("#4", err.Error())
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
		//	TODO: handle error
		log.Println("#1", err)
		return
	}
	go func() {
		_, err := io.Copy(stdinPipe, session)
		if err != nil {
			//	TODO: handle error
			log.Println("#2", err)
		}
		_ = stdinPipe.Close()
	}()

	err = cmd.Start()
	if err != nil {
		//	TODO: handle error
		log.Println("#3", err)
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
				log.Println("signal error: ", err)
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
		//	TODO: handle error
		log.Println("exit error: ", err)
	}
	log.Println(session.RawCommand(), " command exited successfully")
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

func (s *Server) getShell() string {
	out, err := exec.Command("sh", "-c", "grep '^[^#]' /etc/shells").Output()
	if err != nil {
		return "sh"
	}

	if strings.Contains(string(out), "/bin/bash") {
		return "/bin/bash"
	}

	if strings.Contains(string(out), "/usr/bin/bash") {
		return "/usr/bin/bash"
	}

	if strings.Contains(string(out), "/bin/zsh") {
		return "/bin/zsh"
	}

	if strings.Contains(string(out), "/usr/bin/zsh") {
		return "/usr/bin/zsh"
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
		log.Printf("sftp server init error: %s\n", err)
		return
	}
	if err := server.Serve(); err == io.EOF {
		server.Close()
		fmt.Println("sftp client exited session.")
	} else if err != nil {
		fmt.Println("sftp server completed with error:", err)
	}
}
