//go:build windows

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ssh

import (
	"os"
	"syscall"

	"github.com/UserExistsError/conpty"
	"github.com/gliderlabs/ssh"
	log "github.com/sirupsen/logrus"
)

var (
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	setConsoleWindowInfo = kernel32.NewProc("SetConsoleWindowInfo")
)

func Start(cmd interface{}) *conpty.ConPty {
	if shell, ok := cmd.(string); ok {
		f, err := conpty.Start(shell)
		if err != nil {
			log.Errorf("Unable to start ConPTY: %v", err)
			return nil
		}
		return f
	}
	return nil
}

func SetPtySize(f interface{}, win ssh.Window) {
	if cpty, ok := f.(*conpty.ConPty); ok {
		cpty.Resize(win.Width, win.Height)
	} else {
		log.Errorf("Unable to resize ConPTY")
	}
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
