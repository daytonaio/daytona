// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"context"
	"errors"
	"os"
	"syscall"
	"time"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/shirou/gopsutil/v4/process"
)

func (s *SessionService) Delete(ctx context.Context, sessionId string) error {
	session, ok := s.sessions.Get(sessionId)
	if !ok {
		return common_errors.NewNotFoundError(errors.New("session not found"))
	}

	// Terminate process group first with signals (SIGTERM -> SIGKILL)
	err := s.terminateSession(ctx, session)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to terminate session", "sessionId", session.id, "error", err)
		// Continue with cleanup even if termination fails
	}

	// Cancel context after termination
	session.cancel()

	// Clean up session directory
	err = os.RemoveAll(session.Dir(s.configDir))
	if err != nil {
		return common_errors.NewBadRequestError(err)
	}

	s.sessions.Remove(session.id)
	return nil
}

func (s *SessionService) terminateSession(ctx context.Context, session *session) error {
	if session.cmd == nil || session.cmd.Process == nil {
		return nil
	}

	pid := session.cmd.Process.Pid

	_ = s.signalProcessTree(pid, syscall.SIGTERM)

	err := session.cmd.Process.Signal(syscall.SIGTERM)
	if err != nil {
		// If SIGTERM fails, try SIGKILL immediately
		s.logger.WarnContext(ctx, "SIGTERM failed for session, trying SIGKILL", "sessionId", session.id, "error", err)
		_ = s.signalProcessTree(pid, syscall.SIGKILL)
		return session.cmd.Process.Kill()
	}

	// Wait for graceful termination
	if s.waitForTermination(ctx, pid, s.terminationGracePeriod, s.terminationCheckInterval) {
		s.logger.DebugContext(ctx, "Session terminated gracefully", "sessionId", session.id)
		return nil
	}

	s.logger.DebugContext(ctx, "Session timeout, sending SIGKILL to process tree", "sessionId", session.id)
	_ = s.signalProcessTree(pid, syscall.SIGKILL)
	return session.cmd.Process.Kill()
}

func (s *SessionService) signalProcessTree(pid int, sig syscall.Signal) error {
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
		_ = s.signalProcessTree(childPid, sig)
	}

	for _, child := range descendants {
		// Convert to OS process to send custom signal
		if childProc, err := os.FindProcess(int(child.Pid)); err == nil {
			_ = childProc.Signal(sig)
		}
	}

	return nil
}

func (s *SessionService) waitForTermination(ctx context.Context, pid int, timeout, interval time.Duration) bool {
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
