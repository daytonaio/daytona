//go:build windows

// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
)

// ComputerUse implements the Windows variant of the IComputerUse plugin.
//
// Unlike the Linux implementation (which manages Xvfb/xfce4/x11vnc/novnc
// process trees inside the sandbox), the Windows variant has no internal
// processes to supervise — the Windows desktop is always available and any
// VNC/RDP server is managed externally by the sandbox image.
//
// Process-management methods (Initialize/Start/Stop/...) therefore return
// no-op success or trivially-true status. Input methods (mouse, keyboard),
// screenshot, and display info live in their respective *_windows.go files
// in this same package. Accessibility methods (UI Automation) live in
// accessibility_windows.go.
//
// This plugin is spawned by the daemon-side manager (apps/daemon/pkg/toolbox/
// computeruse/manager/manager_windows.go), which uses WTSQueryUserToken +
// CreateProcessAsUser to launch the binary into the active console session
// so that user32 input APIs (SendInput, SetCursorPos) can drive the
// interactive desktop in WinSta0\Default.
type ComputerUse struct{}

// Ensure ComputerUse implements IComputerUse at compile time.
var _ computeruse.IComputerUse = &ComputerUse{}

// ── Process management ──────────────────────────────────────────────────────
// On Windows the "processes" managed by this plugin are conceptually just the
// plugin itself; there is no Xvfb/xfce4 stack to bring up or down.

// Initialize is a no-op on Windows. The desktop is always available.
func (c *ComputerUse) Initialize() (*computeruse.Empty, error) {
	return new(computeruse.Empty), nil
}

// Start is a no-op on Windows. The daemon-side HTTP handler is responsible
// for spawning this plugin via the manager (which is what realises /start),
// so by the time Start() is invoked the plugin is already up.
func (c *ComputerUse) Start() (*computeruse.Empty, error) {
	return new(computeruse.Empty), nil
}

// Stop is a no-op on Windows. The daemon-side HTTP handler kills the plugin
// process via the manager; this method exists to satisfy the interface.
func (c *ComputerUse) Stop() (*computeruse.Empty, error) {
	return new(computeruse.Empty), nil
}

// GetProcessStatus returns the status of the (single) plugin process. If this
// method is being called, the plugin is by definition running.
func (c *ComputerUse) GetProcessStatus() (map[string]computeruse.ProcessStatus, error) {
	return map[string]computeruse.ProcessStatus{
		"computer-use-plugin": {
			Running:     true,
			Priority:    1,
			AutoRestart: false,
			Pid:         nil,
		},
	}, nil
}

// IsProcessRunning reports true for the plugin's own name; false otherwise.
func (c *ComputerUse) IsProcessRunning(req *computeruse.ProcessRequest) (bool, error) {
	return req.ProcessName == "computer-use-plugin", nil
}

// RestartProcess is a no-op on Windows. To restart the plugin, callers should
// invoke /computeruse/stop followed by /computeruse/start, which exercises
// the daemon-side spawn machinery rather than letting the plugin self-restart.
func (c *ComputerUse) RestartProcess(req *computeruse.ProcessRequest) (*computeruse.Empty, error) {
	return new(computeruse.Empty), nil
}

// GetProcessLogs returns an empty string on Windows. The plugin's own logs
// are emitted to stderr (captured by go-plugin's hclog), not to per-process
// log files.
func (c *ComputerUse) GetProcessLogs(req *computeruse.ProcessRequest) (string, error) {
	return "", nil
}

// GetProcessErrors returns an empty string on Windows for the same reason.
func (c *ComputerUse) GetProcessErrors(req *computeruse.ProcessRequest) (string, error) {
	return "", nil
}

// ── Status ──────────────────────────────────────────────────────────────────

// GetStatus reports "active" — if this plugin is being called, computer-use
// is available. (External VNC server presence is not checked here; the
// daemon-side status endpoint can layer that in if needed.)
func (c *ComputerUse) GetStatus() (*computeruse.ComputerUseStatusResponse, error) {
	return &computeruse.ComputerUseStatusResponse{
		Status: "active",
	}, nil
}
