//go:build windows

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ssh

import (
	"os"
	"syscall"
	"unsafe"

	"github.com/gliderlabs/ssh"
)

var (
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	setConsoleWindowInfo = kernel32.NewProc("SetConsoleWindowInfo")
)

func SetPtySize(f *os.File, win ssh.Window) {
	handle := f.Fd()
	var rect struct {
		Left, Top, Right, Bottom int16
	}
	rect.Right = int16(win.Width - 1)
	rect.Bottom = int16(win.Height - 1)

	setConsoleWindowInfo.Call(uintptr(handle), uintptr(1), uintptr(unsafe.Pointer(&rect)))
}

func OsSignalFrom(sig ssh.Signal) os.Signal {
	switch sig {
	case ssh.SIGINT:
		return os.Interrupt
	case ssh.SIGTERM:
		return os.Kill
	case ssh.SIGKILL:
		return os.Kill
	default:
		return os.Kill
	}
}
