// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

// ParseSSHCommand parses the SSH command string returned by the API
// Expected formats:
// - "ssh token@host" (port 22)
// - "ssh -p port token@host"
func ParseSSHCommand(sshCommand string) ([]string, error) {
	parts := strings.Fields(sshCommand)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid SSH command format: %s", sshCommand)
	}

	// Skip the "ssh" part
	args := parts[1:]

	return args, nil
}

// ExecuteSSH runs the SSH command with proper terminal handling
func ExecuteSSH(sshArgs []string) error {
	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		return fmt.Errorf("ssh not found in PATH: %w", err)
	}

	// Create the command
	sshCmd := exec.Command(sshPath, sshArgs...)
	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr

	// Handle signals - forward them to the SSH process
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGWINCH)

	// Start the SSH process
	if err := sshCmd.Start(); err != nil {
		return fmt.Errorf("failed to start SSH: %w", err)
	}

	// Forward signals to the SSH process
	go func() {
		for sig := range sigChan {
			if sshCmd.Process != nil {
				sshCmd.Process.Signal(sig)
			}
		}
	}()

	// Wait for SSH to complete
	err = sshCmd.Wait()

	// Stop signal handling
	signal.Stop(sigChan)
	close(sigChan)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return err
	}

	return nil
}
