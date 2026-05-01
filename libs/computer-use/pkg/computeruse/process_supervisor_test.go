// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestStartAllProcessesWaitsForReadinessBeforeAdvancing(t *testing.T) {
	sleep := lookPath(t, "sleep")
	var (
		mu    sync.Mutex
		order []string
	)

	c := &ComputerUse{processes: map[string]*Process{}}
	c.processes["xfce4"] = readyTestProcess("xfce4", sleep, 200, true, func(ctx context.Context, p *Process) error {
		if !getProcessStatus(p).Running {
			return fmt.Errorf("process is not running")
		}
		mu.Lock()
		defer mu.Unlock()
		if strings.Join(order, ",") != "xvfb" {
			return fmt.Errorf("xfce4 started before xvfb was ready: %v", order)
		}
		order = append(order, "xfce4")
		return nil
	})
	c.processes["xvfb"] = readyTestProcess("xvfb", sleep, 100, true, func(ctx context.Context, p *Process) error {
		if !getProcessStatus(p).Running {
			return fmt.Errorf("process is not running")
		}
		if getProcessStatus(c.processes["xfce4"]).Running {
			return fmt.Errorf("xfce4 started before xvfb was ready")
		}
		mu.Lock()
		defer mu.Unlock()
		order = append(order, "xvfb")
		return nil
	})

	if err := c.startAllProcesses(context.Background()); err != nil {
		t.Fatalf("startAllProcesses() error = %v", err)
	}
	defer c.Stop()

	mu.Lock()
	defer mu.Unlock()
	if got := strings.Join(order, ","); got != "xvfb,xfce4" {
		t.Fatalf("readiness order = %s, want xvfb,xfce4", got)
	}
}

func TestStartAllProcessesReturnsRequiredReadinessError(t *testing.T) {
	sleep := lookPath(t, "sleep")
	c := &ComputerUse{
		processes: map[string]*Process{
			"xvfb": readyTestProcess("xvfb", sleep, 100, true, func(context.Context, *Process) error {
				return fmt.Errorf("display unavailable")
			}),
		},
	}
	c.processes["xvfb"].ReadyName = "fake-display"
	c.processes["xvfb"].ReadyTimeout = 20 * time.Millisecond
	defer c.Stop()

	err := c.startAllProcesses(context.Background())
	if err == nil {
		t.Fatal("startAllProcesses() error = nil, want required readiness error")
	}
	for _, want := range []string{"xvfb", "fake-display", "20ms", "display unavailable"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q should contain %q", err, want)
		}
	}
}

func TestStartAllProcessesDoesNotBlockOnOptionalReadinessFailure(t *testing.T) {
	sleep := lookPath(t, "sleep")
	c := &ComputerUse{processes: map[string]*Process{}}
	c.processes["atspi"] = readyTestProcess("atspi", sleep, 100, false, func(context.Context, *Process) error {
		return fmt.Errorf("a11y bus unavailable")
	})
	c.processes["atspi"].ReadyTimeout = 20 * time.Millisecond

	requiredReady := make(chan struct{})
	c.processes["x11vnc"] = readyTestProcess("x11vnc", sleep, 200, true, func(ctx context.Context, p *Process) error {
		if !getProcessStatus(p).Running {
			return fmt.Errorf("process is not running")
		}
		close(requiredReady)
		return nil
	})
	defer c.Stop()

	if err := c.startAllProcesses(context.Background()); err != nil {
		t.Fatalf("startAllProcesses() error = %v", err)
	}

	select {
	case <-requiredReady:
	case <-time.After(time.Second):
		t.Fatal("required process did not start after optional readiness failure")
	}
}

func TestRestartLoopClosesProcessFilesEachIteration(t *testing.T) {
	falsePath := lookPath(t, "false")
	var closes atomic.Int32
	previousOpenProcessFile := openProcessFile
	openProcessFile = func(string) (processFile, error) {
		return &countingProcessFile{closes: &closes}, nil
	}
	defer func() { openProcessFile = previousOpenProcessFile }()

	c := &ComputerUse{restartDelay: time.Millisecond}
	process := &Process{
		Name:        "xvfb",
		Command:     falsePath,
		AutoRestart: true,
		LogFile:     "xvfb.log",
		ErrFile:     "xvfb.err",
	}

	go c.startProcess(process)
	waitUntil(t, time.Second, func() bool {
		return closes.Load() >= 4
	})
	c.stopProcess(process)
}

type countingProcessFile struct {
	closes *atomic.Int32
}

func (f *countingProcessFile) Write(p []byte) (int, error) {
	return len(p), nil
}

func (f *countingProcessFile) Close() error {
	f.closes.Add(1)
	return nil
}

func readyTestProcess(name, command string, priority int, required bool, ready func(context.Context, *Process) error) *Process {
	return &Process{
		Name:         name,
		Command:      command,
		Args:         []string{"30"},
		Priority:     priority,
		Required:     required,
		AutoRestart:  true,
		Ready:        ready,
		ReadyName:    name + "-ready",
		ReadyTimeout: time.Second,
	}
}

func lookPath(t *testing.T, name string) string {
	t.Helper()
	path, err := exec.LookPath(name)
	if err != nil {
		t.Fatalf("%s not found: %v", name, err)
	}
	return path
}

func waitUntil(t *testing.T, timeout time.Duration, ready func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if ready() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition did not become true before timeout")
}
