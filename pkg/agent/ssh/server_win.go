//go:build windows

// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/UserExistsError/conpty"
	"github.com/gliderlabs/ssh"
	log "github.com/sirupsen/logrus"
)

var (
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	setConsoleWindowInfo = kernel32.NewProc("SetConsoleWindowInfo")
)

func Start(cmd interface{}) (*conpty.ConPty, error) {
	if shell, ok := cmd.(*exec.Cmd); ok {
		f, err := conpty.Start(`c:\windows\system32\cmd.exe`, conpty.ConPtyEnv(shell.Env), conpty.ConPtyWorkDir(shell.Dir))
		if err != nil {
			return nil, fmt.Errorf("Unable to start ConPTY: %v", err)
		}
		return f, nil
	}
	return nil, fmt.Errorf("Unable to start ConPTY")
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
