//go:build linux

// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
)

func SpawnTTY(opts SpawnTTYOptions) error {
	shell := GetShell()
	cmd := exec.Command(shell)

	cmd.Dir = opts.Dir

	cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", opts.Term))
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, fmt.Sprintf("SHELL=%s", shell))
	cmd.Env = append(cmd.Env, opts.Env...)

	var f *os.File
	var err error
	if opts.InitCols >= 1 && opts.InitRows >= 1 {
		f, err = pty.StartWithSize(cmd, &pty.Winsize{Rows: uint16(opts.InitRows), Cols: uint16(opts.InitCols)})
	} else {
		f, err = pty.Start(cmd)
	}
	if err != nil {
		return err
	}

	defer f.Close()

	go func() {
		for win := range opts.SizeCh {
			syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
				uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(win.Height), uint16(win.Width), 0, 0})))
		}
	}()

	go func() {
		io.Copy(f, opts.StdIn) // stdin
	}()

	_, err = io.Copy(opts.StdOut, f) // stdout
	return err
}
