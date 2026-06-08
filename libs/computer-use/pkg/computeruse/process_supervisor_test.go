// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
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

func TestRestartProcessWaitsForReadiness(t *testing.T) {
	computerUse := &ComputerUse{processes: make(map[string]*Process)}
	process := blockingHelperProcess("x11vnc", 300)
	release := make(chan struct{})
	events := make(chan string, 1)
	process.readinessProbe = gatedReadinessProbe(process.Name, process, events, release)
	process.readinessTimeout = 5 * time.Second
	computerUse.processes[process.Name] = process
	t.Cleanup(func() { _, _ = computerUse.Stop() })

	errCh := make(chan error, 1)
	go func() {
		_, err := computerUse.RestartProcess(&toolbox.ProcessRequest{ProcessName: process.Name})
		errCh <- err
	}()

	expectReadinessEvent(t, events, process.Name)
	assertNoRestartResult(t, errCh, 100*time.Millisecond)
	close(release)

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("RestartProcess() error = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("RestartProcess() did not return")
	}
}

func TestRestartProcessStopsProcessOnReadinessFailure(t *testing.T) {
	sentinel := errors.New("vnc not ready")
	computerUse := &ComputerUse{processes: make(map[string]*Process)}
	process := blockingHelperProcess("x11vnc", 300)
	failReadiness := make(chan struct{})
	events := make(chan string, 1)
	process.readinessProbe = gatedFailingReadinessProbe(process.Name, process, events, failReadiness, sentinel)
	process.readinessTimeout = 300 * time.Millisecond
	computerUse.processes[process.Name] = process
	t.Cleanup(func() { _, _ = computerUse.Stop() })

	errCh := make(chan error, 1)
	go func() {
		_, err := computerUse.RestartProcess(&toolbox.ProcessRequest{ProcessName: process.Name})
		errCh <- err
	}()

	expectReadinessEvent(t, events, process.Name)
	assertNoRestartResult(t, errCh, 20*time.Millisecond)
	close(failReadiness)

	select {
	case err := <-errCh:
		if !errors.Is(err, sentinel) {
			t.Fatalf("RestartProcess() error = %v, want %v", err, sentinel)
		}
		if err == nil || !strings.Contains(err.Error(), process.Name) {
			t.Fatalf("error = %v, want process name", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("RestartProcess() did not return")
	}
	eventuallyProcessStopped(t, process)
}

func TestStartAllProcessesWaitsForReadinessInPriorityOrder(t *testing.T) {
	computerUse, releases, events := readinessOrderedComputerUse()
	t.Cleanup(func() { _, _ = computerUse.Stop() })

	errCh := make(chan error, 1)
	go func() { errCh <- computerUse.startAllProcesses() }()

	expectReadinessEvent(t, events, "xvfb")
	assertNoReadinessEvent(t, events, 100*time.Millisecond)
	close(releases["xvfb"])

	expectReadinessEvent(t, events, "xfce4")
	assertNoReadinessEvent(t, events, 100*time.Millisecond)
	close(releases["xfce4"])

	expectReadinessEvent(t, events, "x11vnc")
	assertNoReadinessEvent(t, events, 100*time.Millisecond)
	close(releases["x11vnc"])

	expectReadinessEvent(t, events, "novnc")
	close(releases["novnc"])

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("startAllProcesses() error = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("startAllProcesses() did not return")
	}
}

func TestStartAllProcessesStopsStartedProcessesOnReadinessFailure(t *testing.T) {
	sentinel := errors.New("desktop not ready")
	computerUse, releases, events := readinessOrderedComputerUse()
	failReadiness := make(chan struct{})
	computerUse.processes["xfce4"].readinessProbe = gatedFailingReadinessProbe("xfce4", computerUse.processes["xfce4"], events, failReadiness, sentinel)
	computerUse.processes["xfce4"].readinessTimeout = 300 * time.Millisecond
	t.Cleanup(func() { _, _ = computerUse.Stop() })

	errCh := make(chan error, 1)
	go func() { errCh <- computerUse.startAllProcesses() }()

	expectReadinessEvent(t, events, "xvfb")
	close(releases["xvfb"])
	expectReadinessEvent(t, events, "xfce4")
	assertNoReadinessEvent(t, events, 20*time.Millisecond)
	close(failReadiness)

	select {
	case err := <-errCh:
		if !errors.Is(err, sentinel) {
			t.Fatalf("startAllProcesses() error = %v, want %v", err, sentinel)
		}
		if err == nil || !strings.Contains(err.Error(), "xfce4") {
			t.Fatalf("error = %v, want process name", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("startAllProcesses() did not return")
	}

	if processMarkedRunning(computerUse.processes["x11vnc"]) {
		t.Fatal("x11vnc should not start after xfce4 failure")
	}
	if processMarkedRunning(computerUse.processes["novnc"]) {
		t.Fatal("novnc should not start after xfce4 failure")
	}
	eventuallyProcessStopped(t, computerUse.processes["xvfb"])
	eventuallyProcessStopped(t, computerUse.processes["xfce4"])
}

func TestStartAllProcessesDoesNotUseFixedStartupSleepWhenReadinessSucceeds(t *testing.T) {
	computerUse, _ := immediateReadinessComputerUseWithCalls()
	t.Cleanup(func() { _, _ = computerUse.Stop() })

	startedAt := time.Now()
	if err := computerUse.startAllProcesses(); err != nil {
		t.Fatalf("startAllProcesses() error = %v", err)
	}
	if elapsed := time.Since(startedAt); elapsed > 1500*time.Millisecond {
		t.Fatalf("startAllProcesses() took %s, fixed startup sleeps likely remain", elapsed)
	}
}

func TestWaitForReadinessReportsLastError(t *testing.T) {
	sentinel := errors.New("still not ready")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := waitForReadiness(ctx, func(ctx context.Context) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("waitForReadiness() error = %v, want %v", err, sentinel)
	}
}

func TestXDisplaySocketPath(t *testing.T) {
	tests := []struct {
		display string
		path    string
		ok      bool
	}{
		{display: ":0", path: "/tmp/.X11-unix/X0", ok: true},
		{display: ":12.0", path: "/tmp/.X11-unix/X12", ok: true},
		{display: "localhost:10.0", ok: false},
		{display: ":bad", ok: false},
	}

	for _, tt := range tests {
		path, ok := xDisplaySocketPath(tt.display)
		if ok != tt.ok || path != tt.path {
			t.Fatalf("xDisplaySocketPath(%q) = %q, %v; want %q, %v", tt.display, path, ok, tt.path, tt.ok)
		}
	}
}

func TestXfce4WindowManagerReady(t *testing.T) {
	if err := xfce4WindowManagerReady("_NET_SUPPORTING_WM_CHECK(WINDOW): window id # 0x200003"); err != nil {
		t.Fatalf("xfce4WindowManagerReady() error = %v", err)
	}
	if err := xfce4WindowManagerReady("_NET_SUPPORTING_WM_CHECK:  not found."); err == nil {
		t.Fatal("xfce4WindowManagerReady() should reject missing window manager property")
	}
}

func TestTCPReadinessProbe(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	defer listener.Close()

	accepted := make(chan struct{})
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			_ = conn.Close()
		}
		close(accepted)
	}()

	tcpAddr := listener.Addr().(*net.TCPAddr)
	if err := tcpReadinessProbe("127.0.0.1", fmt.Sprint(tcpAddr.Port))(context.Background()); err != nil {
		t.Fatalf("tcpReadinessProbe() error = %v", err)
	}

	select {
	case <-accepted:
	case <-time.After(time.Second):
		t.Fatal("listener did not accept readiness connection")
	}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_COMPUTER_USE_HELPER_PROCESS") != "1" {
		return
	}
	if os.Getenv("GO_COMPUTER_USE_HELPER_MODE") == "block" {
		select {}
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

type desktopProcessSpec struct {
	name     string
	priority int
}

var desktopProcessSpecs = []desktopProcessSpec{
	{name: "xvfb", priority: 100},
	{name: "xfce4", priority: 200},
	{name: "x11vnc", priority: 300},
	{name: "novnc", priority: 400},
}

func readinessOrderedComputerUse() (*ComputerUse, map[string]chan struct{}, chan string) {
	computerUse := &ComputerUse{processes: make(map[string]*Process)}
	releases := make(map[string]chan struct{})
	events := make(chan string, 4)
	for _, spec := range desktopProcessSpecs {
		process := blockingHelperProcess(spec.name, spec.priority)
		release := make(chan struct{})
		process.readinessProbe = gatedReadinessProbe(spec.name, process, events, release)
		process.readinessTimeout = 5 * time.Second
		computerUse.processes[spec.name] = process
		releases[spec.name] = release
	}
	return computerUse, releases, events
}

func immediateReadinessComputerUseWithCalls() (*ComputerUse, map[string]*atomic.Bool) {
	calls := make(map[string]*atomic.Bool)
	computerUse := &ComputerUse{processes: make(map[string]*Process)}
	for _, spec := range desktopProcessSpecs {
		process := blockingHelperProcess(spec.name, spec.priority)
		called := &atomic.Bool{}
		process.readinessProbe = trackedImmediateReadinessProbe(process, called)
		process.readinessTimeout = time.Second
		calls[spec.name] = called
		computerUse.processes[spec.name] = process
	}
	return computerUse, calls
}

func blockingHelperProcess(name string, priority int) *Process {
	return &Process{
		Name:        name,
		Command:     os.Args[0],
		Args:        []string{"-test.run=TestHelperProcess"},
		Priority:    priority,
		AutoRestart: false,
		Env: map[string]string{
			"GO_WANT_COMPUTER_USE_HELPER_PROCESS": "1",
			"GO_COMPUTER_USE_HELPER_MODE":         "block",
		},
	}
}

func gatedReadinessProbe(name string, process *Process, events chan<- string, release <-chan struct{}) func(context.Context) error {
	var sent atomic.Bool
	return func(ctx context.Context) error {
		if !processMarkedRunning(process) {
			return fmt.Errorf("process %s is not running", name)
		}
		if sent.CompareAndSwap(false, true) {
			select {
			case events <- name:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func gatedFailingReadinessProbe(name string, process *Process, events chan<- string, fail <-chan struct{}, err error) func(context.Context) error {
	var sent atomic.Bool
	return func(ctx context.Context) error {
		if !processMarkedRunning(process) {
			return fmt.Errorf("process %s is not running", name)
		}
		if sent.CompareAndSwap(false, true) {
			select {
			case events <- name:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		select {
		case <-fail:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func trackedImmediateReadinessProbe(process *Process, called *atomic.Bool) func(context.Context) error {
	return func(context.Context) error {
		if !processMarkedRunning(process) {
			return fmt.Errorf("process %s is not running", process.Name)
		}
		called.Store(true)
		return nil
	}
}

func processMarkedRunning(process *Process) bool {
	process.mu.Lock()
	defer process.mu.Unlock()
	return process.running
}

func eventuallyProcessStopped(t *testing.T, process *Process) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if !processMarkedRunning(process) {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("process %s is still running", process.Name)
}

func expectReadinessEvent(t *testing.T, events <-chan string, want string) {
	t.Helper()
	select {
	case got := <-events:
		if got != want {
			t.Fatalf("readiness event = %s, want %s", got, want)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for readiness event %s", want)
	}
}

func assertNoReadinessEvent(t *testing.T, events <-chan string, wait time.Duration) {
	t.Helper()
	select {
	case got := <-events:
		t.Fatalf("unexpected readiness event %s", got)
	case <-time.After(wait):
	}
}

func assertNoRestartResult(t *testing.T, results <-chan error, wait time.Duration) {
	t.Helper()
	select {
	case err := <-results:
		t.Fatalf("RestartProcess() returned before readiness completed: %v", err)
	case <-time.After(wait):
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
