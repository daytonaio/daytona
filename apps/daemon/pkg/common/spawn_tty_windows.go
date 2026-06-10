//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync"
	"syscall"

	"github.com/UserExistsError/conpty"
)

func SpawnTTY(opts SpawnTTYOptions) error {
	shell := GetShell()

	// conpty.Start passes this verbatim as lpCommandLine to CreateProcessW
	// with lpApplicationName=nil, so an unquoted path containing spaces
	// (e.g. DAYTONA_SHELL=pwsh.exe resolving under C:\Program Files) would
	// be parsed ambiguously (CWE-428). Quote it like NewShellCommand does.
	cmdLine := syscall.EscapeArg(shell)
	if IsPowerShell(shell) {
		cmdLine += " -NoLogo"
	}

	cols := opts.InitCols
	if cols < 1 {
		cols = defaultTTYCols
	}
	rows := opts.InitRows
	if rows < 1 {
		rows = defaultTTYRows
	}

	cptyOpts := []conpty.ConPtyOption{
		conpty.ConPtyDimensions(cols, rows),
	}
	if opts.Dir != "" {
		cptyOpts = append(cptyOpts, conpty.ConPtyWorkDir(opts.Dir))
	}
	if len(opts.Env) > 0 {
		// ConPtyEnv replaces the child's entire environment block, so
		// append the extras to the inherited environment to keep the
		// Linux semantics of Env.
		cptyOpts = append(cptyOpts, conpty.ConPtyEnv(append(os.Environ(), opts.Env...)))
	}

	cpty, err := conpty.Start(cmdLine, cptyOpts...)
	if err != nil {
		slog.Error("Failed to start ConPTY", "command", cmdLine, "error", err)
		return err
	}

	// conpty.Close is not idempotent: it closes raw Windows handles
	// (including the process handle), so a second call could close
	// unrelated handles that reused the same values.
	var closeOnce sync.Once
	closePty := func() {
		closeOnce.Do(func() {
			if err := cpty.Close(); err != nil {
				slog.Debug("Failed to close ConPTY", "error", err)
			}
		})
	}
	defer closePty()

	go func() {
		for win := range opts.SizeCh {
			if err := cpty.Resize(win.Width, win.Height); err != nil {
				slog.Debug("Failed to resize ConPTY", "error", err)
			}
		}
	}()

	go func() {
		if _, err := io.Copy(cpty, opts.StdIn); err != nil && err != io.EOF {
			slog.Debug("ConPTY stdin copy error", "error", err)
		}
	}()

	// Wait for the shell in the background. On natural exit, close the
	// pty: ClosePseudoConsole detaches conhost, so the stdout copy below
	// drains the remaining output and hits EOF instead of blocking forever.
	waitErrCh := make(chan error, 1)
	go func() {
		exitCode, waitErr := cpty.Wait(context.Background())
		if waitErr == nil {
			slog.Debug("ConPTY session exited", "exit_code", exitCode)
		}
		waitErrCh <- waitErr
		closePty()
	}()

	// Mirror the Linux flow: the stdout copy is the synchronization point.
	// It returns when the client goes away (write error) or when the wait
	// goroutine closed the pty after the shell exited (read error/EOF).
	if _, err := io.Copy(opts.StdOut, cpty); err != nil && err != io.EOF {
		slog.Debug("ConPTY stdout copy error", "error", err)
	}

	// Client disconnect: terminate the shell (ClosePseudoConsole tears down
	// the attached console processes) and unblock the wait goroutine, whose
	// polled process handle is now closed.
	closePty()

	if waitErr := <-waitErrCh; waitErr != nil {
		slog.Debug("ConPTY wait error", "error", waitErr)
		return waitErr
	}
	return nil
}
