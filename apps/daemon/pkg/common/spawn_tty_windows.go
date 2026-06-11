//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"context"
	"io"
	"log/slog"

	"github.com/UserExistsError/conpty"
)

const (
	defaultTTYCols = 80
	defaultTTYRows = 24
)

type TTYSize struct {
	Height int
	Width  int
}

type SpawnTTYOptions struct {
	// Ctx, when non-nil, bounds the spawned session: on cancellation the
	// TTY is torn down and the attached process is terminated.
	Ctx      context.Context
	Dir      string
	StdIn    io.Reader
	StdOut   io.Writer
	Term     string
	Env      []string
	InitCols int
	InitRows int
	SizeCh   <-chan TTYSize
}

func SpawnTTY(opts SpawnTTYOptions) error {
	shell := GetShell()

	cmdLine := shell
	if IsPowerShell(shell) {
		cmdLine = shell + " -NoLogo"
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

	cpty, err := conpty.Start(cmdLine, cptyOpts...)
	if err != nil {
		slog.Error("Failed to start ConPTY", "command", cmdLine, "error", err)
		return err
	}
	defer cpty.Close()

	// Bind the session lifetime to the caller's context so a disconnected
	// client unblocks the Wait below and the deferred Close kills the
	// attached process instead of leaking it.
	parent := opts.Ctx
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	go func() {
		for win := range opts.SizeCh {
			if err := cpty.Resize(win.Width, win.Height); err != nil {
				slog.Debug("Failed to resize ConPTY", "error", err)
			}
		}
	}()

	go func() {
		// Client EOF/disconnect ends the stdin copy: cancel so Wait returns.
		defer cancel()
		if _, err := io.Copy(cpty, opts.StdIn); err != nil && err != io.EOF {
			slog.Debug("ConPTY stdin copy error", "error", err)
		}
	}()

	go func() {
		if _, err := io.Copy(opts.StdOut, cpty); err != nil && err != io.EOF {
			slog.Debug("ConPTY stdout copy error", "error", err)
		}
	}()

	exitCode, err := cpty.Wait(ctx)
	if err != nil {
		if ctx.Err() != nil {
			// Cancellation is a normal session end: the caller's ctx fired
			// (client connection gone) or stdin hit EOF. The deferred Close
			// tears down the ConPTY and terminates the attached process.
			slog.Debug("ConPTY session ended by client disconnect")
			return nil
		}
		slog.Debug("ConPTY wait error", "error", err)
		return err
	}

	slog.Debug("ConPTY session exited", "exit_code", exitCode)
	return nil
}
