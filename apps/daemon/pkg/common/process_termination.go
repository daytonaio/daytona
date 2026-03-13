// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"context"
	"log/slog"
	"os"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v4/process"
)

// TerminateProcessTreeGracefully attempts graceful termination with SIGTERM,
// waits for the grace period, then forcefully kills with SIGKILL if needed.
// If the process is a process group leader, signals the entire group.
// Otherwise, recursively walks and signals the process tree.
func TerminateProcessTreeGracefully(ctx context.Context, logger *slog.Logger, proc *os.Process, gracePeriod, checkInterval time.Duration) error {
	if proc == nil || proc.Pid <= 0 {
		return nil
	}

	pid := proc.Pid

	// Check if this process is a process group leader (PGID == PID)
	// If so, we can signal the entire group efficiently
	isGroupLeader := isProcessGroupLeader(pid)

	// Send SIGTERM (graceful)
	var err error
	if isGroupLeader {
		// Signal entire process group
		err = syscall.Kill(-pid, syscall.SIGTERM)
	} else {
		// Walk tree and signal each process
		_ = signalProcessTree(pid, syscall.SIGTERM)
		err = proc.Signal(syscall.SIGTERM)
	}

	// Wait for graceful termination
	if err == nil && waitForTermination(ctx, pid, gracePeriod, checkInterval) {
		logger.DebugContext(ctx, "PID terminated gracefully", "PID", pid)
		return nil
	}

	// Timeout reached or SIGTERM failed, force kill with SIGKILL
	logger.WarnContext(ctx, "PID SIGTERM timed out or failed, sending SIGKILL", "PID", pid)
	if isGroupLeader {
		// Kill entire process group
		return syscall.Kill(-pid, syscall.SIGKILL)
	}
	// Walk tree and kill each process
	_ = signalProcessTree(pid, syscall.SIGKILL)
	return proc.Kill()
}

// isProcessGroupLeader checks if a process is a process group leader (PGID == PID)
func isProcessGroupLeader(pid int) bool {
	pgid, err := syscall.Getpgid(pid)
	if err != nil {
		return false
	}
	return pgid == pid
}

// signalProcessTree recursively sends a signal to a process and all its descendants
func signalProcessTree(pid int, sig syscall.Signal) error {
	parent, err := process.NewProcess(int32(pid))
	if err != nil {
		return err
	}

	descendants, err := parent.Children()
	if err != nil {
		return err
	}

	for _, child := range descendants {
		childPid := int(child.Pid)
		// Recurse depth-first before signaling this child
		_ = signalProcessTree(childPid, sig)
		// Convert to OS process to send custom signal
		childProc, err := os.FindProcess(childPid)
		if err == nil {
			_ = childProc.Signal(sig)
		}
	}

	return nil
}

// waitForTermination waits for a process and its children to terminate
func waitForTermination(ctx context.Context, pid int, timeout, interval time.Duration) bool {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return false
		case <-ticker.C:
			_, err := process.NewProcess(int32(pid))
			if err != nil {
				// Process doesn't exist anymore
				return true
			}
		}
	}
}
