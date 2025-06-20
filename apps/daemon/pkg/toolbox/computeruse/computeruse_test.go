// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"testing"
)

func TestNewComputerUse(t *testing.T) {
	computerUse := NewComputerUse()
	if computerUse == nil {
		t.Fatal("NewComputerUse() returned nil")
	}

	if computerUse.processes == nil {
		t.Fatal("processes map is nil")
	}

	if len(computerUse.processes) != 0 {
		t.Fatalf("Expected 0 processes, got %d", len(computerUse.processes))
	}
}

func TestInitializeProcesses(t *testing.T) {
	computerUse := NewComputerUse()

	// Set up config directory
	computerUse.configDir = "/tmp/test-computeruse"

	// Initialize processes
	computerUse.initializeProcesses()

	// Check that all expected processes are created
	expectedProcesses := []string{"xvfb", "xfce4", "x11vnc", "novnc"}
	for _, processName := range expectedProcesses {
		process, exists := computerUse.processes[processName]
		if !exists {
			t.Errorf("Process %s not found", processName)
			continue
		}

		if process.Name != processName {
			t.Errorf("Expected process name %s, got %s", processName, process.Name)
		}

		if !process.AutoRestart {
			t.Errorf("Process %s should have auto-restart enabled", processName)
		}
	}

	// Check priorities are set correctly
	expectedPriorities := map[string]int{
		"xvfb":   100,
		"xfce4":  200,
		"x11vnc": 300,
		"novnc":  400,
	}

	for processName, expectedPriority := range expectedPriorities {
		process := computerUse.processes[processName]
		if process.Priority != expectedPriority {
			t.Errorf("Process %s expected priority %d, got %d",
				processName, expectedPriority, process.Priority)
		}
	}

	// Check that xfce4 has the correct environment variables
	xfce4 := computerUse.processes["xfce4"]
	if xfce4.Env["DISPLAY"] != ":1" {
		t.Errorf("Expected DISPLAY=:1, got %s", xfce4.Env["DISPLAY"])
	}
	if xfce4.Env["HOME"] != "/home/agent" {
		t.Errorf("Expected HOME=/home/agent, got %s", xfce4.Env["HOME"])
	}
	if xfce4.Env["USER"] != "agent" {
		t.Errorf("Expected USER=agent, got %s", xfce4.Env["USER"])
	}
}

func TestGetProcessesByPriority(t *testing.T) {
	computerUse := NewComputerUse()
	computerUse.configDir = "/tmp/test-computeruse"
	computerUse.initializeProcesses()

	processes := computerUse.getProcessesByPriority()

	// Check that processes are sorted by priority (ascending)
	for i := 0; i < len(processes)-1; i++ {
		if processes[i].Priority > processes[i+1].Priority {
			t.Errorf("Processes not sorted by priority: %s (%d) comes before %s (%d)",
				processes[i].Name, processes[i].Priority,
				processes[i+1].Name, processes[i+1].Priority)
		}
	}

	// Check that we have all 4 processes
	if len(processes) != 4 {
		t.Errorf("Expected 4 processes, got %d", len(processes))
	}
}

func TestIsProcessRunning(t *testing.T) {
	computerUse := NewComputerUse()
	computerUse.configDir = "/tmp/test-computeruse"
	computerUse.initializeProcesses()

	// Initially, no processes should be running
	if computerUse.IsProcessRunning("xvfb") {
		t.Error("Xvfb should not be running initially")
	}

	// Test with non-existent process
	if computerUse.IsProcessRunning("nonexistent") {
		t.Error("Non-existent process should not be running")
	}
}

func TestGetProcessStatus(t *testing.T) {
	computerUse := NewComputerUse()
	computerUse.configDir = "/tmp/test-computeruse"
	computerUse.initializeProcesses()

	status := computerUse.GetProcessStatus()

	// Check that status contains all processes
	expectedProcesses := []string{"xvfb", "xfce4", "x11vnc", "novnc"}
	for _, processName := range expectedProcesses {
		processStatusInterface, exists := status[processName]
		if !exists {
			t.Errorf("Status for process %s not found", processName)
			continue
		}

		// Type assert to map[string]interface{}
		processStatus, ok := processStatusInterface.(map[string]interface{})
		if !ok {
			t.Errorf("Status for %s is not a map[string]interface{}", processName)
			continue
		}

		// Check that status contains expected fields
		if _, ok := processStatus["running"]; !ok {
			t.Errorf("Status for %s missing 'running' field", processName)
		}
		if _, ok := processStatus["priority"]; !ok {
			t.Errorf("Status for %s missing 'priority' field", processName)
		}
		if _, ok := processStatus["autoRestart"]; !ok {
			t.Errorf("Status for %s missing 'autoRestart' field", processName)
		}

		// Initially, all processes should not be running
		if processStatus["running"] != false {
			t.Errorf("Process %s should not be running initially", processName)
		}
	}
}

func TestRestartProcess(t *testing.T) {
	computerUse := NewComputerUse()
	computerUse.configDir = "/tmp/test-computeruse"
	computerUse.initializeProcesses()

	// Test restarting non-existent process
	err := computerUse.RestartProcess("nonexistent")
	if err == nil {
		t.Error("Expected error when restarting non-existent process")
	}

	// Test restarting existing process (should not error, but process won't actually start)
	err = computerUse.RestartProcess("xvfb")
	if err != nil {
		t.Errorf("Unexpected error when restarting xvfb: %v", err)
	}
}

func TestGetProcessLogs(t *testing.T) {
	computerUse := NewComputerUse()
	computerUse.configDir = "/tmp/test-computeruse"
	computerUse.initializeProcesses()

	// Test getting logs for non-existent process
	_, err := computerUse.GetProcessLogs("nonexistent")
	if err == nil {
		t.Error("Expected error when getting logs for non-existent process")
	}

	// Test getting logs for process without log file
	_, err = computerUse.GetProcessLogs("xvfb")
	if err == nil {
		t.Error("Expected error when getting logs for process without log file")
	}
}

func TestGetProcessErrors(t *testing.T) {
	computerUse := NewComputerUse()
	computerUse.configDir = "/tmp/test-computeruse"
	computerUse.initializeProcesses()

	// Test getting errors for non-existent process
	_, err := computerUse.GetProcessErrors("nonexistent")
	if err == nil {
		t.Error("Expected error when getting errors for non-existent process")
	}

	// Test getting errors for process without error file
	_, err = computerUse.GetProcessErrors("xvfb")
	if err == nil {
		t.Error("Expected error when getting errors for process without error file")
	}
}

// Benchmark tests
func BenchmarkNewComputerUse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewComputerUse()
	}
}

func BenchmarkGetProcessStatus(b *testing.B) {
	computerUse := NewComputerUse()
	computerUse.configDir = "/tmp/test-computeruse"
	computerUse.initializeProcesses()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		computerUse.GetProcessStatus()
	}
}

func BenchmarkIsProcessRunning(b *testing.B) {
	computerUse := NewComputerUse()
	computerUse.configDir = "/tmp/test-computeruse"
	computerUse.initializeProcesses()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		computerUse.IsProcessRunning("xvfb")
	}
}
