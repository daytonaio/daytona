// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"testing"
	"time"

	toolbox "github.com/daytonaio/daemon/pkg/toolbox/computeruse"
)

type processFileCounts struct {
	opens    atomic.Int32
	closes   atomic.Int32
	logOpens atomic.Int32
	errOpens atomic.Int32
}

type countingWriteCloser struct {
	closes *atomic.Int32
}

func (c countingWriteCloser) Write(p []byte) (int, error) {
	return len(p), nil
}

func (c countingWriteCloser) Close() error {
	c.closes.Add(1)
	return nil
}

func TestStartProcessClosesLogFilesEachRestart(t *testing.T) {
	computerUse, process, counts := supervisedHelperProcess(10 * time.Millisecond)

	go computerUse.startProcess(process)
	t.Cleanup(func() { _ = computerUse.stopProcess(process) })

	waitForAtomicAtLeast(t, &counts.closes, 4, 5*time.Second)
	if err := computerUse.stopProcess(process); err != nil {
		t.Fatalf("stopProcess() error = %v", err)
	}

	if counts.opens.Load() != counts.closes.Load() {
		t.Fatalf("opened process files = %d, closed = %d", counts.opens.Load(), counts.closes.Load())
	}
	if counts.logOpens.Load() == 0 {
		t.Fatal("helper log file was not opened")
	}
	if counts.errOpens.Load() == 0 {
		t.Fatal("helper error file was not opened")
	}
}

func TestStopProcessPreventsAutoRestart(t *testing.T) {
	computerUse, process, counts := supervisedHelperProcess(25 * time.Millisecond)

	go computerUse.startProcess(process)
	t.Cleanup(func() { _ = computerUse.stopProcess(process) })

	waitForAtomicAtLeast(t, &counts.closes, 2, 5*time.Second)
	if err := computerUse.stopProcess(process); err != nil {
		t.Fatalf("stopProcess() error = %v", err)
	}

	closedAfterStop := counts.closes.Load()
	openedAfterStop := counts.opens.Load()
	time.Sleep(3 * computerUse.getProcessRestartDelay())

	if counts.closes.Load() != closedAfterStop {
		t.Fatalf("process files closed after stop = %d, want %d", counts.closes.Load(), closedAfterStop)
	}
	if counts.opens.Load() != openedAfterStop {
		t.Fatalf("process files opened after stop = %d, want %d", counts.opens.Load(), openedAfterStop)
	}
}

func TestRestartProcessWaitsForOldSupervisor(t *testing.T) {
	computerUse, process, counts := supervisedHelperProcess(time.Hour)
	computerUse.processes = map[string]*Process{process.Name: process}

	go computerUse.startProcess(process)
	t.Cleanup(func() { _ = computerUse.stopProcess(process) })
	waitForAtomicAtLeast(t, &counts.closes, 2, 5*time.Second)

	_, err := computerUse.RestartProcess(&toolbox.ProcessRequest{ProcessName: process.Name})
	if err != nil {
		t.Fatalf("RestartProcess() error = %v", err)
	}

	waitForAtomicAtLeast(t, &counts.closes, 4, 5*time.Second)

	process.mu.Lock()
	running := process.running
	process.mu.Unlock()
	if !running {
		t.Fatal("process supervisor should be running after restart")
	}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_COMPUTER_USE_HELPER_PROCESS") != "1" {
		return
	}
	fmt.Fprintln(os.Stdout, "stdout")
	fmt.Fprintln(os.Stderr, "stderr")
	os.Exit(0)
}

func supervisedHelperProcess(restartDelay time.Duration) (*ComputerUse, *Process, *processFileCounts) {
	counts := &processFileCounts{}
	computerUse := &ComputerUse{
		openProcessFile: func(name string, flag int, perm os.FileMode) (io.WriteCloser, error) {
			counts.opens.Add(1)
			switch name {
			case "helper.log":
				counts.logOpens.Add(1)
			case "helper.err":
				counts.errOpens.Add(1)
			}
			return countingWriteCloser{closes: &counts.closes}, nil
		},
		processRestartDelay: restartDelay,
	}
	return computerUse, helperProcess(), counts
}

func helperProcess() *Process {
	return &Process{
		Name:        "helper",
		Command:     os.Args[0],
		Args:        []string{"-test.run=TestHelperProcess"},
		AutoRestart: true,
		Env: map[string]string{
			"GO_WANT_COMPUTER_USE_HELPER_PROCESS": "1",
		},
		LogFile: "helper.log",
		ErrFile: "helper.err",
	}
}

func waitForAtomicAtLeast(t *testing.T, value *atomic.Int32, want int32, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if value.Load() >= want {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("value = %d, want at least %d", value.Load(), want)
}
