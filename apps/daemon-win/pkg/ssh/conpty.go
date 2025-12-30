// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build windows

package ssh

import (
	"context"
	"io"

	"github.com/UserExistsError/conpty"
	"github.com/daytonaio/daemon-win/pkg/common"
	"github.com/gliderlabs/ssh"
	log "github.com/sirupsen/logrus"
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

// SpawnConPTY spawns a Windows ConPTY pseudo-terminal
func SpawnConPTY(opts SpawnConPTYOptions) error {
	shell := common.GetShell()

	// Build command line - PowerShell needs -NoExit for interactive mode
	cmdLine := shell
	if common.IsPowerShell(shell) {
		cmdLine = shell + " -NoLogo -NoExit"
	}

	// Create ConPTY with initial size and working directory
	cptyOpts := []conpty.ConPtyOption{
		conpty.ConPtyDimensions(int(opts.Cols), int(opts.Rows)),
	}
	if opts.Dir != "" {
		cptyOpts = append(cptyOpts, conpty.ConPtyWorkDir(opts.Dir))
	}

	cpty, err := conpty.Start(cmdLine, cptyOpts...)
	if err != nil {
		log.Errorf("Failed to start ConPTY with command '%s': %v", cmdLine, err)
		return err
	}
	defer cpty.Close()

	// Handle window resize events
	go func() {
		for win := range opts.WinCh {
			if err := cpty.Resize(int(win.Width), int(win.Height)); err != nil {
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
