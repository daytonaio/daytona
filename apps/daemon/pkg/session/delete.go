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
	"github.com/daytonaio/daemon/pkg/childreap"
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

	// Signal the whole process group (negative pid). Falls back to the tree
	// walker for any descendants that escaped the group via setpgid/setsid.
	_ = syscall.Kill(-pid, syscall.SIGTERM)
	_ = s.signalProcessTree(pid, syscall.SIGTERM)

	// Wait for graceful termination
	if s.waitForTermination(ctx, pid, s.terminationGracePeriod, s.terminationCheckInterval) {
		s.logger.DebugContext(ctx, "Session terminated gracefully", "sessionId", session.id)
		s.reapSession(session)
		return nil
	}

	s.logger.DebugContext(ctx, "Session timeout, sending SIGKILL to process group", "sessionId", session.id)
	_ = syscall.Kill(-pid, syscall.SIGKILL)
	_ = s.signalProcessTree(pid, syscall.SIGKILL)
	err := session.cmd.Process.Kill()
	s.reapSession(session)
	return err
}

// reapSession runs cmd.Wait in the background to release parent-side
// pipe descriptors and update cmd.ProcessState. Zombie collection itself
// is handled by the PID-1 reaper installed in pkg/childreap, so we
// don't need to block the Delete request path on this call — detaching
// it keeps the DELETE response fast (HTTP clients with ~10s timeouts
// can't afford 5s+ tails in the cleanup path).
//
// Uses Reap (not Wait) because there are no Stdout/Stderr buffers to
// drain — session.cmd's std{in,out,err} are *os.File pipes (or unset).
func (s *SessionService) reapSession(session *session) {
	if session.cmd == nil {
		return
	}
	go func() {
		_, _ = childreap.Reap(session.cmd)
	}()
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
