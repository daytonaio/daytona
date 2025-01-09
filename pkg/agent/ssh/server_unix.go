//go:build !windows

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ssh

import (
	"os"
	"syscall"
	"unsafe"

	"github.com/gliderlabs/ssh"
)

func SetPtySize(f *os.File, win ssh.Window) {
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(win.Height), uint16(win.Width), 0, 0})))
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
