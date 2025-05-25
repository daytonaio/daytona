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

type TTYSize struct {
	Height int
	Width  int
}

type SpawnTTYOptions struct {
	Dir    string
	StdIn  io.Reader
	StdOut io.Writer
	Term   string
	Env    []string
	SizeCh <-chan TTYSize
}

func SpawnTTY(opts SpawnTTYOptions) error {
	shell := GetShell()
	cmd := exec.Command(shell)

	cmd.Dir = opts.Dir

	cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", opts.Term))
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, fmt.Sprintf("SHELL=%s", shell))
	cmd.Env = append(cmd.Env, opts.Env...)

	f, err := pty.Start(cmd)
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
