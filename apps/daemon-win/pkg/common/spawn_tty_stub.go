//go:build !windows

// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"errors"
	"io"
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
	return errors.New("SpawnTTY is only available on Windows")
}
