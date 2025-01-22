//go:build !windows

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ssh

import (
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

func Start(cmd interface{}) *os.File {
	if command, ok := cmd.(*exec.Cmd); ok {
		f, err := pty.Start(command)
		if err != nil {
			log.Errorf("Unable to start PTY: %v", err)
			return nil
		}
		return f
	}
	return nil
}

func SetPtySize(f interface{}, win ssh.Window) {
	if file, ok := f.(*os.File); ok {
		syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), uintptr(syscall.TIOCSWINSZ),
			uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(win.Height), uint16(win.Width), 0, 0})))
	} else {
		log.Errorf("Unable to resize PTY")
	}
}

func OsSignalFrom(sig ssh.Signal) os.Signal {
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
