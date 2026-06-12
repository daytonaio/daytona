// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import "io"

const (
	defaultTTYCols = 80
	defaultTTYRows = 24
)

// TTYSize is a terminal resize event delivered via SpawnTTYOptions.SizeCh.
type TTYSize struct {
	Height int
	Width  int
}

// SpawnTTYOptions is the cross-platform contract for SpawnTTY; the two
// build-tagged SpawnTTY implementations (pty on Linux, ConPTY on Windows)
// share this single option set so callers compile identically on both.
type SpawnTTYOptions struct {
	// Dir is the working directory of the spawned shell.
	Dir string
	// StdIn and StdOut carry the terminal byte streams.
	StdIn  io.Reader
	StdOut io.Writer
	// Term sets the TERM environment variable of the shell on Linux.
	// Ignored on Windows: ConPTY defines the VT semantics itself and
	// Windows shells do not consult TERM.
	Term string
	// Env entries are appended to the inherited process environment.
	Env []string
	// InitCols and InitRows set the initial terminal dimensions. When
	// either is < 1, Windows falls back to defaultTTYCols x defaultTTYRows
	// (ConPTY requires explicit dimensions) while Linux leaves the pty at
	// its default size until the first SizeCh event.
	InitCols int
	InitRows int
	// SizeCh delivers resize events. The sender must close it once no more
	// resizes will be sent, or the per-session resize goroutine leaks.
	SizeCh <-chan TTYSize
}
