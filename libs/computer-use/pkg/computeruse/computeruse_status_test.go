// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"os"
	"path/filepath"
	"testing"

	toolbox "github.com/daytonaio/daemon/pkg/toolbox/computeruse"
)

func TestAtspiStatusUsesA11yHealth(t *testing.T) {
	healthCalls := 0
	c := &ComputerUse{
		processes: map[string]*Process{
			"atspi": {Name: "atspi", Priority: 250, AutoRestart: false},
		},
		a11yHealth: func() bool {
			healthCalls++
			return true
		},
	}

	status, err := c.GetProcessStatus()
	if err != nil {
		t.Fatalf("GetProcessStatus() error = %v", err)
	}
	if !status["atspi"].Running {
		t.Fatal("atspi status should use the AT-SPI health check")
	}
	if status["atspi"].Pid != nil {
		t.Fatal("atspi status should not report the launcher PID")
	}
	if status["atspi"].AutoRestart {
		t.Fatal("atspi should be a one-shot bootstrap process")
	}

	running, err := c.IsProcessRunning(&toolbox.ProcessRequest{ProcessName: "atspi"})
	if err != nil {
		t.Fatalf("IsProcessRunning() error = %v", err)
	}
	if !running {
		t.Fatal("IsProcessRunning(atspi) should use the AT-SPI health check")
	}
	if healthCalls != 2 {
		t.Fatalf("health check calls = %d, want 2", healthCalls)
	}
}

func TestInitializeProcessesRegistersAtspiAsBootstrap(t *testing.T) {
	binDir := t.TempDir()
	atspiPath := filepath.Join(binDir, "at-spi-bus-launcher")
	if err := os.WriteFile(atspiPath, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatalf("write fake at-spi-bus-launcher: %v", err)
	}

	t.Setenv("PATH", binDir)
	t.Setenv("DBUS_SESSION_BUS_ADDRESS", "")

	c := &ComputerUse{
		processes: make(map[string]*Process),
		configDir: t.TempDir(),
	}
	c.initializeProcesses(t.TempDir())

	atspi, ok := c.processes["atspi"]
	if !ok {
		t.Fatal("atspi process should be registered when at-spi-bus-launcher is available")
	}
	if atspi.AutoRestart {
		t.Fatal("atspi launcher should not auto-restart")
	}
}
