// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ssh

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
	"github.com/daytonaio/daytona/pkg/agent/config"
	"github.com/gliderlabs/ssh"
	"github.com/pkg/sftp"
	crypto_ssh "golang.org/x/crypto/ssh"
	"golang.org/x/sys/unix"
)

func Start() {
	forwardedTCPHandler := &ssh.ForwardedTCPHandler{}

	sshServer := ssh.Server{
		Addr: ":2222",
		Handler: func(s ssh.Session) {
			switch ss := s.Subsystem(); ss {
			case "":
			case "sftp":
				sftpHandler(s)
				return
			default:
				fmt.Fprintf(s, "Subsystem %s not supported\n", ss)
				s.Exit(1)
				return
			}

			ptyReq, winCh, isPty := s.Pty()
			if s.RawCommand() == "" && isPty {
				handlePty(s, ptyReq, winCh)
			} else {
				handleNonPty(s)
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
			"sftp": sftpHandler,
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
	log.Fatal(sshServer.ListenAndServe())
}

func handlePty(s ssh.Session, ptyReq ssh.Pty, winCh <-chan ssh.Window) {
	c, err := config.GetConfig()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	shell := getShell()
	cmd := exec.Command(shell)

	cmd.Dir = c.ProjectDir

	if ssh.AgentRequested(s) {
		l, err := ssh.NewAgentListener()
		if err != nil {
			fmt.Println("#5", err.Error())
		}
		defer l.Close()
		go ssh.ForwardAgentConnections(l, s)
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
		io.Copy(f, s) // stdin
	}()
	io.Copy(s, f) // stdout
}

func handleNonPty(s ssh.Session) {
	args := []string{}
	if len(s.Command()) > 0 {
		args = append([]string{"-c"}, s.RawCommand())
	}

	fmt.Println("args: ", args)

	c, err := config.GetConfig()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	cmd := exec.Command("/bin/sh", args...)

	cmd.Env = append(cmd.Env, os.Environ()...)

	if ssh.AgentRequested(s) {
		l, err := ssh.NewAgentListener()
		if err != nil {
			fmt.Println("#4", err.Error())
		}
		defer l.Close()
		go ssh.ForwardAgentConnections(l, s)
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", "SSH_AUTH_SOCK", l.Addr().String()))
	}

	cmd.Dir = c.ProjectDir
	cmd.Stdout = s
	cmd.Stderr = s.Stderr()
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		//	TODO: handle error
		log.Println("#1", err)
		return
	}
	go func() {
		_, err := io.Copy(stdinPipe, s)
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
	s.Signals(sigs)
	defer func() {
		s.Signals(nil)
		close(sigs)
	}()
	go func() {
		for sig := range sigs {
			signal := osSignalFrom(sig)
			cmd.Process.Signal(signal)
		}
	}()
	err = cmd.Wait()

	if err != nil {
		log.Println(s.RawCommand(), " ", err)
		s.Exit(127)
		return
	}

	err = s.Exit(0)
	if err != nil {
		//	TODO: handle error
		log.Println("exit error: ", err)
	}
	log.Println(s.RawCommand(), " command exited successfully")
}

func osSignalFrom(sig ssh.Signal) os.Signal {
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

func getShell() string {
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

func sftpHandler(sess ssh.Session) {
	debugStream := io.Discard
	serverOptions := []sftp.ServerOption{
		sftp.WithDebug(debugStream),
	}
	server, err := sftp.NewServer(
		sess,
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
