// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build unix

package computeruse

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	toolbox "github.com/daytonaio/daemon/pkg/toolbox/computeruse"
)

func TestRunProcessOnceClosesLogFilesEachRun(t *testing.T) {
	restoreFileLimit := lowerOpenFileLimit(t, 64)
	defer restoreFileLimit()

	computerUse := &ComputerUse{}
	process := helperProcess(t, "exit")

	// Each run opens stdout and stderr log files. With the old supervisor leak,
	// 80 runs would leave 160 extra descriptors open and hit this lowered limit.

	for i := 0; i < 80; i++ {
		if err := computerUse.runProcessOnce(context.Background(), process); err != nil {
			t.Fatalf("runProcessOnce() iteration %d error = %v", i+1, err)
		}
	}
}

func TestStopProcessWaitsForSupervisorExit(t *testing.T) {
	computerUse := &ComputerUse{}
	process := helperProcess(t, "block")

	go computerUse.startProcess(process)
	t.Cleanup(func() { _ = computerUse.stopProcess(process) })
	waitForProcessPID(t, process, 5*time.Second)

	if err := computerUse.stopProcess(process); err != nil {
		t.Fatalf("stopProcess() error = %v", err)
	}
	assertProcessStopped(t, process)
}

func TestStopProcessInterruptsRestartBackoff(t *testing.T) {
	computerUse := &ComputerUse{}
	process := helperProcess(t, "short")

	go computerUse.startProcess(process)
	t.Cleanup(func() { _ = computerUse.stopProcess(process) })
	waitForRestartBackoff(t, process, 5*time.Second)

	startedAt := time.Now()
	if err := computerUse.stopProcess(process); err != nil {
		t.Fatalf("stopProcess() error = %v", err)
	}
	if elapsed := time.Since(startedAt); elapsed > time.Second {
		t.Fatalf("stopProcess() took %s, want it to interrupt restart delay", elapsed)
	}
	assertProcessStopped(t, process)
}

func TestRestartProcessWaitsForOldSupervisor(t *testing.T) {
	computerUse := &ComputerUse{processes: make(map[string]*Process)}
	process := helperProcess(t, "block")
	computerUse.processes[process.Name] = process

	go computerUse.startProcess(process)
	t.Cleanup(func() { _ = computerUse.stopProcess(process) })
	oldPID := waitForProcessPID(t, process, 5*time.Second)

	_, err := computerUse.RestartProcess(&toolbox.ProcessRequest{ProcessName: process.Name})
	if err != nil {
		t.Fatalf("RestartProcess() error = %v", err)
	}

	newPID := waitForDifferentProcessPID(t, process, oldPID, 5*time.Second)
	if newPID == oldPID {
		t.Fatalf("process PID after restart = %d, want a new PID", newPID)
	}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_COMPUTER_USE_HELPER_PROCESS") != "1" {
		return
	}
	fmt.Fprintln(os.Stdout, "stdout")
	fmt.Fprintln(os.Stderr, "stderr")
	switch os.Getenv("GO_COMPUTER_USE_HELPER_MODE") {
	case "block":
		time.Sleep(time.Minute)
	case "short":
		time.Sleep(100 * time.Millisecond)
	}
	os.Exit(0)
}

func helperProcess(t *testing.T, mode string) *Process {
	t.Helper()

	dir := t.TempDir()
	return &Process{
		Name:        "helper",
		Command:     os.Args[0],
		Args:        []string{"-test.run=TestHelperProcess"},
		AutoRestart: true,
		Env: map[string]string{
			"GO_WANT_COMPUTER_USE_HELPER_PROCESS": "1",
			"GO_COMPUTER_USE_HELPER_MODE":         mode,
		},
		LogFile: filepath.Join(dir, "helper.log"),
		ErrFile: filepath.Join(dir, "helper.err"),
	}
}

func lowerOpenFileLimit(t *testing.T, fdHeadroom uint64) func() {
	t.Helper()

	var old syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &old); err != nil {
		t.Skipf("cannot read file descriptor limit: %v", err)
	}

	openFDs, err := os.ReadDir("/dev/fd")
	if err != nil {
		t.Skipf("cannot count open file descriptors: %v", err)
	}

	target := uint64(len(openFDs)) + fdHeadroom
	if target < 128 {
		target = 128
	}
	if old.Cur <= target {
		return func() {}
	}

	next := old
	next.Cur = target
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &next); err != nil {
		t.Skipf("cannot lower file descriptor limit: %v", err)
	}

	return func() {
		if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &old); err != nil {
			t.Errorf("failed to restore file descriptor limit: %v", err)
		}
	}
}

func waitForProcessPID(t *testing.T, process *Process, timeout time.Duration) int {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if pid := processPID(process); pid > 0 {
			return pid
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("process did not start")
	return 0
}

func waitForRestartBackoff(t *testing.T, process *Process, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	sawChild := false
	for time.Now().Before(deadline) {
		process.mu.Lock()
		running := process.running
		cmd := process.cmd
		process.mu.Unlock()
		if cmd != nil {
			sawChild = true
		}
		if sawChild && running && cmd == nil {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("process did not enter restart backoff")
}

func waitForDifferentProcessPID(t *testing.T, process *Process, oldPID int, timeout time.Duration) int {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if pid := processPID(process); pid > 0 && pid != oldPID {
			return pid
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("process did not restart with a new PID; old PID = %d", oldPID)
	return 0
}

func processPID(process *Process) int {
	process.mu.Lock()
	defer process.mu.Unlock()

	if !process.running || process.cmd == nil || process.cmd.Process == nil {
		return 0
	}
	return process.cmd.Process.Pid
}

func assertProcessStopped(t *testing.T, process *Process) {
	t.Helper()

	process.mu.Lock()
	defer process.mu.Unlock()
	if process.running {
		t.Fatal("process is still marked running")
	}
	if process.cmd != nil {
		t.Fatal("process command was not cleared")
	}
	if process.done != nil {
		t.Fatal("process done channel was not cleared")
	}
}
