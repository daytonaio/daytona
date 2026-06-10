//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

// Package childreap is a thin os/exec wrapper on Windows. The Linux build
// wraps go-reaper to recover the exit status of orphaned children when the
// daemon runs as PID 1; Windows has no PID-1 reparenting and no SIGCHLD, so
// no reaper is installed here. The helpers still honor the (exitCode, err)
// contract shared with Linux: a non-zero exit folds into the returned exit
// code with a nil error, and a non-nil error is returned only when cmd
// couldn't be started or no exit status could be recovered. Callers must
// check exitCode != 0 to detect command failure.
//
// Timing, however, is NOT at parity with Linux: Wait and Reap both
// delegate to raw cmd.Wait, which blocks until the runtime's pipe-drain
// copy goroutines finish (golang/go#23019). In particular, Windows Reap
// does not deliver the Linux Reap guarantee of returning as soon as the
// exit status is known — with piped stdio (StdoutPipe, bytes.Buffer, …)
// it can block indefinitely on an undrained or inherited pipe unless
// cmd.WaitDelay is set. Current Windows callers avoid this by using only
// *os.File stdio (no copy goroutines); teardown-path callers that pipe
// stdio MUST set cmd.WaitDelay.
package childreap

import (
	"errors"
	"os/exec"
)

func Start() {}

// Wait waits for a started cmd and returns its exit code. See the Linux
// implementation for the full (exitCode, err) contract.
func Wait(cmd *exec.Cmd) (int, error) {
	if cmd == nil || cmd.Process == nil {
		return -1, errors.New("childreap.Wait: cmd not started")
	}
	err := cmd.Wait()
	// exec.ErrWaitDelay substitutes for a nil error only: the process itself
	// exited (ProcessState is populated, and a non-zero status would have
	// surfaced as *exec.ExitError instead) but an inherited I/O pipe was
	// still open when WaitDelay expired. The exit status WAS recovered, so
	// honor the package contract and return it without an error.
	if errors.Is(err, exec.ErrWaitDelay) {
		return cmd.ProcessState.ExitCode(), nil
	}
	return foldExitError(err)
}

// Reap is equivalent to Wait on Windows: with no PID-1 reaper to race
// against, cmd.Wait is always the sole status consumer. Unlike the Linux
// Reap, it does NOT return as soon as the exit status is known: it is raw
// cmd.Wait and may block indefinitely draining stdio pipes unless
// cmd.WaitDelay is set (see the package comment).
func Reap(cmd *exec.Cmd) (int, error) {
	return Wait(cmd)
}

// foldExitError mirrors the Linux contract: *exec.ExitError carries a real
// exit status, so it folds into (code, nil); any other error means the
// status could not be recovered.
func foldExitError(waitErr error) (int, error) {
	if waitErr == nil {
		return 0, nil
	}
	var ee *exec.ExitError
	if errors.As(waitErr, &ee) {
		return ee.ExitCode(), nil
	}
	return -1, waitErr
}
