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
		// SpawnTTY owns SizeCh consumption even on failure: senders (the
		// ssh window-change forwarders, the terminal ws reader) block in
		// an unbuffered send, and closing the channel cannot unblock a
		// parked sender. Keep draining until the sender closes the
		// channel, or every failed spawn permanently leaks the sender
		// goroutine.
		if opts.SizeCh != nil {
			go func() {
				for range opts.SizeCh {
				}
			}()
		}
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
	// pty so the stdout copy below unblocks: ClosePseudoConsole detaches
	// conhost, which fails the pending read. Note conpty.Close is
	// all-or-nothing — it also closes cmdOut, the very handle the copy
	// reads from — so output buffered but not yet read when the shell
	// exits may be dropped (the read fails with ERROR_INVALID_HANDLE
	// rather than draining to EOF). Full tail fidelity would need a
	// two-phase close (ClosePseudoConsole, drain, then CloseHandle) that
	// the library does not offer.
	waitCtx, cancelWait := context.WithCancel(context.Background())
	defer cancelWait()
	waitErrCh := make(chan error, 1)
	go func() {
		exitCode, waitErr := cpty.Wait(waitCtx)
		switch {
		case waitErr == nil:
			slog.Debug("ConPTY session exited", "exit_code", exitCode)
			closePty()
		case waitCtx.Err() != nil:
			// Canceled by the disconnect path below — not a wait
			// failure. Report nil so clean disconnects return nil,
			// mirroring Linux.
			waitErr = nil
		default:
			// The wait itself failed (the process handle could not be
			// queried). Close anyway so the stdout copy below cannot
			// block forever on a pty nobody else will close.
			closePty()
		}
		waitErrCh <- waitErr
	}()

	// Mirror the Linux flow: the stdout copy is the synchronization point.
	// It returns when the client goes away (write error) or when the wait
	// goroutine closed the pty after the shell exited (read error/EOF).
	if _, err := io.Copy(opts.StdOut, cpty); err != nil && err != io.EOF {
		slog.Debug("ConPTY stdout copy error", "error", err)
	}

	// Stop the waiter before touching the handles: cancel its context and
	// join it, so it cannot poll the raw process handle after closePty
	// releases it (the handle value may be recycled; see the closeOnce
	// comment above). conpty.Wait checks the context at least once per
	// second, so the join is bounded.
	cancelWait()
	waitErr := <-waitErrCh

	// Client disconnect: terminate the shell (ClosePseudoConsole tears
	// down the attached console processes). No-op when the wait goroutine
	// already closed the pty after a natural exit.
	closePty()

	if waitErr != nil {
		slog.Debug("ConPTY wait error", "error", waitErr)
		return waitErr
	}
	return nil
}
