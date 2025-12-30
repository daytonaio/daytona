//go:build windows

// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"context"
	"io"

	"github.com/UserExistsError/conpty"
	log "github.com/sirupsen/logrus"
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

	// For interactive terminal sessions, we don't use -Command or -NonInteractive
	// Just start the shell directly for a proper interactive experience
	cmdLine := shell
	if IsPowerShell(shell) {
		// Use -NoLogo for cleaner startup, -NoExit to keep the session open
		cmdLine = shell + " -NoLogo"
	}

	// Create ConPTY options
	cptyOpts := []conpty.ConPtyOption{
		conpty.ConPtyDimensions(80, 24), // Default size, will be resized
	}
	if opts.Dir != "" {
		cptyOpts = append(cptyOpts, conpty.ConPtyWorkDir(opts.Dir))
	}

	// Start ConPTY
	cpty, err := conpty.Start(cmdLine, cptyOpts...)
	if err != nil {
		log.Errorf("Failed to start ConPTY: %v", err)
		return err
	}
	defer cpty.Close()

	// Handle window resize
	go func() {
		for win := range opts.SizeCh {
			if err := cpty.Resize(win.Width, win.Height); err != nil {
				log.Debugf("Failed to resize ConPTY: %v", err)
			}
		}
	}()

	// Copy stdin to ConPTY
	go func() {
		_, err := io.Copy(cpty, opts.StdIn)
		if err != nil && err != io.EOF {
			log.Debugf("stdin copy error: %v", err)
		}
	}()

	// Copy ConPTY output to stdout
	go func() {
		_, err := io.Copy(opts.StdOut, cpty)
		if err != nil && err != io.EOF {
			log.Debugf("stdout copy error: %v", err)
		}
	}()

	// Wait for the process to exit
	exitCode, err := cpty.Wait(context.Background())
	if err != nil {
		log.Debugf("ConPTY wait error: %v", err)
		return err
	}

	log.Debugf("ConPTY session exited with code: %d", exitCode)
	return nil
}
