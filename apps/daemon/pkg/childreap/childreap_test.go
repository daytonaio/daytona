// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package childreap

import (
	"os/exec"
	"syscall"
	"testing"
	"time"
)

// TestWait_NormalExit covers the happy path: Wait collects the child via
// cmd.Wait() without any reaper interference.
func TestWait_NormalExit(t *testing.T) {
	cmd := exec.Command("sh", "-c", "exit 0")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	code, err := Wait(cmd)
	if err != nil {
		t.Fatalf("Wait returned unexpected error: %v", err)
	}
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

// TestWait_NonZeroExit confirms Wait returns the real exit code in the
// normal-error branch (cmd.Wait returns *exec.ExitError).
func TestWait_NonZeroExit(t *testing.T) {
	cmd := exec.Command("sh", "-c", "exit 42")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	code, err := Wait(cmd)
	if err != nil {
		t.Fatalf("Wait returned unexpected error: %v", err)
	}
	if code != 42 {
		t.Errorf("expected exit code 42, got %d", code)
	}
}

// TestWait_NotStarted exercises the cmd-was-never-started guard.
func TestWait_NotStarted(t *testing.T) {
	cmd := exec.Command("true")
	// Deliberately do not call cmd.Start.
	_, err := Wait(cmd)
	if err == nil {
		t.Fatalf("expected error when cmd not started")
	}
}

// TestWait_ECHILDRecovery_PendingFirst simulates the reaper winning the
// race: an external wait4 reaps the child before cmd.Wait runs (so
// cmd.Wait will see ECHILD), and the reaper's status is published via
// recordStatus before Wait is called. Wait must read from the pending map.
func TestWait_ECHILDRecovery_PendingFirst(t *testing.T) {
	cmd := exec.Command("sh", "-c", "exit 7")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	pid := cmd.Process.Pid

	// Steal the zombie before cmd.Wait runs. Blocking wait so we know the
	// child has actually exited and we hold its WaitStatus.
	var ws syscall.WaitStatus
	if _, err := syscall.Wait4(pid, &ws, 0, nil); err != nil {
		t.Fatalf("manual Wait4: %v", err)
	}
	// Publish the stolen status into childreap's pending map. Mirrors what
	// the dispatcher does when go-reaper reports a reaped pid that has no
	// registered waiter yet.
	recordStatus(pid, ws)

	code, err := Wait(cmd)
	if err != nil {
		t.Fatalf("Wait returned unexpected error: %v", err)
	}
	if code != 7 {
		t.Errorf("expected recovered exit code 7, got %d", code)
	}
}

// TestWait_ECHILDRecovery_StatusArrivesAfterPark covers the race where the
// reaper publishes the status AFTER Wait has already registered and
// parked on the waiter channel. The dispatcher path should hand the status
// to reg.ch and Wait should return promptly.
func TestWait_ECHILDRecovery_StatusArrivesAfterPark(t *testing.T) {
	cmd := exec.Command("sh", "-c", "exit 9")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	pid := cmd.Process.Pid

	var ws syscall.WaitStatus
	if _, err := syscall.Wait4(pid, &ws, 0, nil); err != nil {
		t.Fatalf("manual Wait4: %v", err)
	}

	// Publish the status only after Wait has had a chance to register and
	// hit ECHILD. The exact delay isn't important; we just need it to
	// happen strictly after Wait's register() call.
	go func() {
		time.Sleep(50 * time.Millisecond)
		recordStatus(pid, ws)
	}()

	code, err := Wait(cmd)
	if err != nil {
		t.Fatalf("Wait returned unexpected error: %v", err)
	}
	if code != 9 {
		t.Errorf("expected recovered exit code 9, got %d", code)
	}
}

// TestWait_ECHILDRecovery_Timeout exercises the failure mode: cmd.Wait
// returned ECHILD and the reaper never published a status. Wait should
// give up after recoveryTimeout and return an error.
//
// Uses a stand-in for cmd.Process.Pid since real wait4 won't return
// ECHILD without prior reaping. We pick a pid that will never be reported
// (a value outside the valid pid range so it can't collide with anything).
func TestWait_ECHILDRecovery_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timeout test in short mode")
	}

	// Save and restore the global timeout so this test runs in
	// reasonable wall time without affecting prod behavior.
	orig := overrideRecoveryTimeoutForTest(100 * time.Millisecond)
	defer overrideRecoveryTimeoutForTest(orig)

	cmd := exec.Command("sh", "-c", "exit 0")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	pid := cmd.Process.Pid

	// Reap externally so cmd.Wait returns ECHILD, but never publish status.
	var ws syscall.WaitStatus
	if _, err := syscall.Wait4(pid, &ws, 0, nil); err != nil {
		t.Fatalf("manual Wait4: %v", err)
	}

	start := time.Now()
	_, err := Wait(cmd)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatalf("expected timeout error, got nil")
	}
	if elapsed < 80*time.Millisecond {
		t.Errorf("Wait returned faster than recovery timeout (%s)", elapsed)
	}
	if elapsed > 500*time.Millisecond {
		t.Errorf("Wait took too long: %s", elapsed)
	}
}
