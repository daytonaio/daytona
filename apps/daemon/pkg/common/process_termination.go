// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"context"
	"os"
	"syscall"
	"time"

	"github.com/daytonaio/daemon/internal/util"
	"github.com/shirou/gopsutil/v4/process"
	log "github.com/sirupsen/logrus"
)

const (
	DefaultGracePeriod   = 5 * time.Second
	DefaultCheckInterval = 100 * time.Millisecond
)

// TerminateProcessTreeGracefully attempts graceful termination with SIGTERM,
// waits for the grace period, then forcefully kills with SIGKILL if needed.
// If the process is a process group leader, signals the entire group.
// Otherwise, recursively walks and signals the process tree.
func TerminateProcessTreeGracefully(ctx context.Context, proc *os.Process, gracePeriod *time.Duration) error {
	if proc == nil || proc.Pid <= 0 {
		return nil
	}

	if gracePeriod == nil {
		gracePeriod = util.Pointer(DefaultGracePeriod)
	}

	pid := proc.Pid

	// Check if this process is a process group leader (PGID == PID)
	// If so, we can signal the entire group efficiently
	isGroupLeader := isProcessGroupLeader(pid)

	// Send SIGTERM (graceful)
	if isGroupLeader {
		// Signal entire process group
		_ = syscall.Kill(-pid, syscall.SIGTERM)
	} else {
		// Walk tree and signal each process
		_ = signalProcessTree(pid, syscall.SIGTERM)
		_ = proc.Signal(syscall.SIGTERM)
	}

	// Wait for graceful termination
	if waitForTermination(ctx, pid, *gracePeriod, DefaultCheckInterval) {
		log.Debugf("PID %d terminated gracefully", pid)
		return nil
	}

	// Timeout reached, force kill with SIGKILL
	log.Debugf("PID %d timeout, sending SIGKILL", pid)
	if isGroupLeader {
		// Kill entire process group
		return syscall.Kill(-pid, syscall.SIGKILL)
	} else {
		// Walk tree and kill each process
		_ = signalProcessTree(pid, syscall.SIGKILL)
		return proc.Kill()
	}
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

	// Recursively signal all descendants
	for _, child := range descendants {
		_ = signalProcessTree(int(child.Pid), sig)
	}

	// Signal all direct children
	for _, child := range descendants {
		childProc, err := os.FindProcess(int(child.Pid))
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
			parent, err := process.NewProcess(int32(pid))
			if err != nil {
				// Process doesn't exist anymore
				return true
			}
			children, err := parent.Children()
			if err != nil {
				// Unable to enumerate children - likely process is dying/dead
				return true
			}
			if len(children) == 0 {
				return true
			}
		}
	}
}
