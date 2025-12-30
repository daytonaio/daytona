// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build !windows

package ssh

import (
	"errors"
	"io"

	"github.com/gliderlabs/ssh"
)

// SpawnConPTYOptions contains options for spawning a ConPTY session
type SpawnConPTYOptions struct {
	Dir    string
	StdIn  io.Reader
	StdOut io.Writer
	Cols   uint16
	Rows   uint16
	WinCh  <-chan ssh.Window
}

// SpawnConPTY is a stub for non-Windows platforms
func SpawnConPTY(opts SpawnConPTYOptions) error {
	return errors.New("ConPTY is only available on Windows")
}
