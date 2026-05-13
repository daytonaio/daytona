// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	toolbox "github.com/daytonaio/daemon/pkg/toolbox/computeruse"
)

func TestSelectNoVNCCommandPrefersSupervisedWebsockify(t *testing.T) {
	dir := t.TempDir()
	websockify := filepath.Join(dir, "websockify")
	if err := os.WriteFile(websockify, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatalf("write fake websockify: %v", err)
	}
	t.Setenv("PATH", dir)

	command, args, err := selectNoVNCCommand("5901", "6080")
	if err != nil {
		t.Fatalf("selectNoVNCCommand() error = %v", err)
	}

	if command != websockify {
		t.Fatalf("command = %q, want %q", command, websockify)
	}
	want := "--web=/usr/share/novnc/ 6080 localhost:5901"
	if got := strings.Join(args, " "); got != want {
		t.Fatalf("args = %q, want %q", got, want)
	}
}

func TestSelectNoVNCCommandReturnsClearErrorWithoutWebsockify(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	_, _, err := selectNoVNCCommand("5901", "6080")
	if err == nil {
		t.Fatal("selectNoVNCCommand() error = nil, want missing websockify error")
	}
	if !strings.Contains(err.Error(), "websockify not found in PATH") {
		t.Fatalf("error = %q, want missing websockify", err)
	}
}

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

func TestStartAllProcessesStopsStartedProcessesOnRequiredReadinessFailure(t *testing.T) {
	sleep := lookPath(t, "sleep")
	c := &ComputerUse{processes: map[string]*Process{}}
	c.processes["xvfb"] = readyTestProcess("xvfb", sleep, 100, true, func(_ context.Context, p *Process) error {
		if !getProcessStatus(p).Running {
			return fmt.Errorf("process is not running")
		}
		return nil
	})
	c.processes["xfce4"] = readyTestProcess("xfce4", sleep, 200, true, func(_ context.Context, p *Process) error {
		if !getProcessStatus(p).Running {
			return fmt.Errorf("process is not running")
		}
		return fmt.Errorf("desktop unavailable")
	})
	c.processes["xfce4"].ReadyTimeout = 250 * time.Millisecond
	defer c.Stop()

	err := c.startAllProcesses(context.Background())
	if err == nil {
		t.Fatal("startAllProcesses() error = nil, want required readiness error")
	}

	for _, process := range c.processes {
		p := process
		waitUntil(t, time.Second, func() bool {
			return !getProcessStatus(p).Running
		})
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

func TestStartAllProcessesAllowsStatusAndStopWhileReadinessPolling(t *testing.T) {
	sleep := lookPath(t, "sleep")
	c := &ComputerUse{processes: map[string]*Process{}}

	readinessPolling := make(chan struct{})
	var readinessOnce sync.Once
	c.processes["xvfb"] = readyTestProcess("xvfb", sleep, 100, true, func(context.Context, *Process) error {
		readinessOnce.Do(func() { close(readinessPolling) })
		if _, err := c.GetProcessStatus(); err != nil {
			return err
		}
		return fmt.Errorf("display still warming up")
	})
	c.processes["xvfb"].ReadyTimeout = 50 * time.Millisecond

	errCh := make(chan error, 1)
	go func() { errCh <- c.startAllProcesses(context.Background()) }()

	select {
	case <-readinessPolling:
	case <-time.After(time.Second):
		t.Fatal("readiness did not start polling")
	}

	if _, err := c.GetProcessStatus(); err != nil {
		t.Fatalf("GetProcessStatus() error = %v", err)
	}
	if _, err := c.IsProcessRunning(&toolbox.ProcessRequest{ProcessName: "xvfb"}); err != nil {
		t.Fatalf("IsProcessRunning() error = %v", err)
	}
	c.stopProcess(c.processes["xvfb"])

	err := <-errCh
	if err == nil {
		t.Fatal("startAllProcesses() error = nil, want readiness timeout")
	}
	if !strings.Contains(err.Error(), "xvfb") {
		t.Fatalf("readiness error %q should name xvfb", err)
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

func TestParseXpropWindowID(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		want    string
		wantErr bool
	}{
		{
			name:   "valid window id",
			output: `_NET_SUPPORTING_WM_CHECK(WINDOW): window id # 0x400001`,
			want:   "0x400001",
		},
		{
			name:    "missing property",
			output:  `_NET_SUPPORTING_WM_CHECK:  not found.`,
			wantErr: true,
		},
		{
			name:    "missing id marker",
			output:  `_NET_SUPPORTING_WM_CHECK(WINDOW): window id`,
			wantErr: true,
		},
		{
			name:    "zero id",
			output:  `_NET_SUPPORTING_WM_CHECK(WINDOW): window id # 0x0`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseXpropWindowID(tt.output)
			if tt.wantErr {
				if err == nil {
					t.Fatal("parseXpropWindowID() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("parseXpropWindowID() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("parseXpropWindowID() = %s, want %s", got, tt.want)
			}
		})
	}
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
