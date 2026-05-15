// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package childreap

import (
	"bytes"
	"os/exec"
	"strings"
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

// TestRun covers the cmd.Run replacement helper.
func TestRun(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		code, err := Run(exec.Command("true"))
		if err != nil {
			t.Fatalf("Run returned unexpected error: %v", err)
		}
		if code != 0 {
			t.Errorf("expected exit code 0, got %d", code)
		}
	})

	t.Run("NonZeroExit", func(t *testing.T) {
		code, err := Run(exec.Command("sh", "-c", "exit 17"))
		if err != nil {
			t.Fatalf("Run returned unexpected error: %v", err)
		}
		if code != 17 {
			t.Errorf("expected exit code 17, got %d", code)
		}
	})

	t.Run("StartFailure", func(t *testing.T) {
		// Bogus binary path. cmd.Start should fail.
		_, err := Run(exec.Command("/this/path/does/not/exist/zzz"))
		if err == nil {
			t.Fatalf("expected start error, got nil")
		}
	})

	t.Run("NilCmd", func(t *testing.T) {
		// Must return an error, not panic.
		_, err := Run(nil)
		if err == nil {
			t.Fatalf("expected error for nil cmd, got nil")
		}
	})
}

// TestCombinedOutput covers the cmd.CombinedOutput replacement helper.
func TestCombinedOutput(t *testing.T) {
	t.Run("CapturesBoth", func(t *testing.T) {
		out, code, err := CombinedOutput(
			exec.Command("sh", "-c", "echo out-line && echo err-line >&2"),
		)
		if err != nil {
			t.Fatalf("CombinedOutput returned unexpected error: %v", err)
		}
		if code != 0 {
			t.Errorf("expected exit code 0, got %d", code)
		}
		s := string(out)
		if !strings.Contains(s, "out-line") {
			t.Errorf("missing stdout in output: %q", s)
		}
		if !strings.Contains(s, "err-line") {
			t.Errorf("missing stderr in output: %q", s)
		}
	})

	t.Run("RejectsPresetStdout", func(t *testing.T) {
		cmd := exec.Command("true")
		cmd.Stdout = &bytes.Buffer{}
		_, _, err := CombinedOutput(cmd)
		if err == nil {
			t.Fatal("expected error when cmd.Stdout was already set")
		}
	})

	t.Run("NilCmd", func(t *testing.T) {
		_, _, err := CombinedOutput(nil)
		if err == nil {
			t.Fatalf("expected error for nil cmd, got nil")
		}
	})
}

// TestOutput covers the cmd.Output replacement helper.
func TestOutput(t *testing.T) {
	t.Run("CapturesStdout", func(t *testing.T) {
		out, code, err := Output(exec.Command("sh", "-c", "echo just-stdout"))
		if err != nil {
			t.Fatalf("Output returned unexpected error: %v", err)
		}
		if code != 0 {
			t.Errorf("expected exit code 0, got %d", code)
		}
		if !strings.Contains(string(out), "just-stdout") {
			t.Errorf("missing stdout: %q", string(out))
		}
	})

	t.Run("RejectsPresetStdout", func(t *testing.T) {
		cmd := exec.Command("true")
		cmd.Stdout = &bytes.Buffer{}
		_, _, err := Output(cmd)
		if err == nil {
			t.Fatal("expected error when cmd.Stdout was already set")
		}
	})

	t.Run("NilCmd", func(t *testing.T) {
		_, _, err := Output(nil)
		if err == nil {
			t.Fatalf("expected error for nil cmd, got nil")
		}
	})
}

// TestWait_StalePendingRejected guards the PID-reuse hardening: a status
// stored in pending more than pendingMaxAge before the caller's register
// must NOT be claimed as the caller's exit status, because it almost
// certainly belongs to a previous process whose pid has since been
// recycled by the kernel.
//
// We simulate this by injecting a stale pending entry for the cmd's pid
// (i.e., with an addedAt timestamp well in the past), reaping the child
// externally so cmd.Wait sees ECHILD, then asserting that Wait recovers
// via the recoveryTimeout path (an error) rather than returning the
// stale status.
func TestWait_StalePendingRejected(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stale pending test in short mode")
	}

	// Tight timeouts so the test completes quickly. pendingMaxAge is set
	// well below the staleness of the injected entry; recoveryTimeout is
	// shortened so the failure path returns within ~100ms.
	origMaxAge := overridePendingMaxAgeForTest(50 * time.Millisecond)
	defer overridePendingMaxAgeForTest(origMaxAge)
	origRecovery := overrideRecoveryTimeoutForTest(100 * time.Millisecond)
	defer overrideRecoveryTimeoutForTest(origRecovery)

	cmd := exec.Command("sh", "-c", "exit 0")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	pid := cmd.Process.Pid

	// Reap externally so cmd.Wait will return ECHILD on its own.
	var ws syscall.WaitStatus
	if _, err := syscall.Wait4(pid, &ws, 0, nil); err != nil {
		t.Fatalf("manual Wait4: %v", err)
	}

	// Inject a stale pending entry for the same pid. addedAt is in the
	// past by 1 second — well past pendingMaxAge (50ms). Imitates what
	// would happen if an earlier (unrelated) process had been reaped
	// holding the same pid before the kernel handed it to our cmd.
	// Construct a "would have been exit 99" WaitStatus by hand.
	// On Linux, ExitStatus() reads (status >> 8) & 0xff for exited
	// children, so writing 99 << 8 gives us a status that would surface
	// as 99 if claimed.
	staleWS := syscall.WaitStatus(99 << 8)
	mu.Lock()
	pending[pid] = pendingStatus{ws: staleWS, addedAt: time.Now().Add(-1 * time.Second)}
	mu.Unlock()

	code, err := Wait(cmd)

	// The stale entry must NOT have been claimed. Wait should fall
	// through to the recoveryTimeout error path.
	if err == nil {
		t.Fatalf("expected error when only a stale pending entry was available, got nil and code=%d", code)
	}
	if code == 99 {
		t.Fatalf("Wait claimed the stale pending entry (code=99) — PID-reuse hardening regressed")
	}

	// And the stale entry must be evicted by the claim attempt so a later
	// caller doesn't keep tripping over it.
	mu.Lock()
	_, stillThere := pending[pid]
	mu.Unlock()
	if stillThere {
		t.Errorf("stale pending entry was not evicted after rejected claim")
	}
}
