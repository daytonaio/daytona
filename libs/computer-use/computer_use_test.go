// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package main

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/daytonaio/daemon/pkg/toolbox/computeruse/manager"
	"github.com/hashicorp/go-hclog"
	hc_plugin "github.com/hashicorp/go-plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClient represents a test client that can communicate with the plugin
type TestClient struct {
	client *hc_plugin.Client
	plugin computeruse.IComputerUse
}

// NewTestClient creates a new test client
func NewTestClient() (*TestClient, error) {
	// Build the plugin
	buildCmd := exec.Command("go", "build", "-o", "test-computer-use", ".")
	buildCmd.Dir = "."
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	if err := buildCmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to build plugin: %v", err)
	}

	// Create plugin client
	client := hc_plugin.NewClient(&hc_plugin.ClientConfig{
		HandshakeConfig: manager.ComputerUseHandshakeConfig,
		Plugins: map[string]hc_plugin.Plugin{
			"daytona-computer-use": &computeruse.ComputerUsePlugin{},
		},
		Cmd:     exec.Command("./test-computer-use"),
		Logger:  hclog.New(&hclog.LoggerOptions{Level: hclog.Error}),
		Managed: true,
	})

	// Connect to the plugin
	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("failed to connect to plugin: %v", err)
	}

	// Get the plugin instance
	raw, err := rpcClient.Dispense("daytona-computer-use")
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("failed to dispense plugin: %v", err)
	}

	plugin, ok := raw.(computeruse.IComputerUse)
	if !ok {
		client.Kill()
		return nil, fmt.Errorf("plugin does not implement IComputerUse interface")
	}

	return &TestClient{
		client: client,
		plugin: plugin,
	}, nil
}

// Close closes the test client
func (tc *TestClient) Close() {
	if tc.client != nil {
		tc.client.Kill()
	}
	// Clean up the test binary
	os.Remove("test-computer-use")
}

// TestComputerUsePlugin runs comprehensive tests on the computer-use plugin
func TestComputerUsePlugin(t *testing.T) {
	// Skip if running in CI or headless environment
	if os.Getenv("CI") == "true" || os.Getenv("DISPLAY") == "" {
		t.Skip("Skipping GUI tests in CI or headless environment")
	}

	client, err := NewTestClient()
	require.NoError(t, err)
	defer client.Close()

	// Test all methods
	t.Run("ProcessManagement", func(t *testing.T) {
		testProcessManagement(t, client.plugin)
	})

	t.Run("ScreenshotMethods", func(t *testing.T) {
		testScreenshotMethods(t, client.plugin)
	})

	t.Run("MouseControlMethods", func(t *testing.T) {
		testMouseControlMethods(t, client.plugin)
	})

	t.Run("KeyboardControlMethods", func(t *testing.T) {
		testKeyboardControlMethods(t, client.plugin)
	})

	t.Run("DisplayInfoMethods", func(t *testing.T) {
		testDisplayInfoMethods(t, client.plugin)
	})

	t.Run("StatusMethod", func(t *testing.T) {
		testStatusMethod(t, client.plugin)
	})
}

// testProcessManagement tests all process management methods
func testProcessManagement(t *testing.T, plugin computeruse.IComputerUse) {
	t.Run("Initialize", func(t *testing.T) {
		_, err := plugin.Initialize()
		assert.NoError(t, err)
	})

	t.Run("Start", func(t *testing.T) {
		_, err := plugin.Start()
		assert.NoError(t, err)

		// Wait a bit for processes to start
		time.Sleep(3 * time.Second)
	})

	t.Run("GetProcessStatus", func(t *testing.T) {
		status, err := plugin.GetProcessStatus()
		assert.NoError(t, err)
		assert.NotNil(t, status)

		// Check that we have the expected processes
		expectedProcesses := []string{"xvfb", "xfce4", "x11vnc", "novnc"}
		for _, processName := range expectedProcesses {
			processStatus, exists := status[processName]
			assert.True(t, exists, "Process %s should exist", processName)
			if exists {
				assert.NotNil(t, processStatus)
				assert.True(t, processStatus.Priority > 0, "Process %s should have priority > 0", processName)
			}
		}
	})

	t.Run("IsProcessRunning", func(t *testing.T) {
		// Test with existing process
		_, err := plugin.IsProcessRunning(&computeruse.ProcessRequest{ProcessName: "xvfb"})
		assert.NoError(t, err)
		// Note: In test environment, processes might not actually be running
		// so we just check that the method doesn't error

		// Test with non-existing process
		_, err = plugin.IsProcessRunning(&computeruse.ProcessRequest{ProcessName: "nonexistent"})
		assert.Error(t, err)
	})

	t.Run("GetProcessLogs", func(t *testing.T) {
		// Test with process that has logs
		logs, err := plugin.GetProcessLogs(&computeruse.ProcessRequest{ProcessName: "xfce4"})
		if err != nil {
			// It's okay if logs don't exist yet
			t.Logf("GetProcessLogs returned error (expected in test): %v", err)
		} else {
			assert.NotNil(t, logs)
		}

		// Test with process that doesn't have logs
		_, err = plugin.GetProcessLogs(&computeruse.ProcessRequest{ProcessName: "xvfb"})
		if err != nil {
			// Expected error for processes without log files
			t.Logf("GetProcessLogs for xvfb returned error (expected): %v", err)
		}
	})

	t.Run("GetProcessErrors", func(t *testing.T) {
		// Test with process that has error logs
		errors, err := plugin.GetProcessErrors(&computeruse.ProcessRequest{ProcessName: "xfce4"})
		if err != nil {
			// It's okay if error logs don't exist yet
			t.Logf("GetProcessErrors returned error (expected in test): %v", err)
		} else {
			assert.NotNil(t, errors)
		}

		// Test with process that doesn't have error logs
		_, err = plugin.GetProcessErrors(&computeruse.ProcessRequest{ProcessName: "xvfb"})
		if err != nil {
			// Expected error for processes without error files
			t.Logf("GetProcessErrors for xvfb returned error (expected): %v", err)
		}
	})

	t.Run("RestartProcess", func(t *testing.T) {
		_, err := plugin.RestartProcess(&computeruse.ProcessRequest{ProcessName: "xvfb"})
		assert.NoError(t, err)

		// Test with non-existing process
		_, err = plugin.RestartProcess(&computeruse.ProcessRequest{ProcessName: "nonexistent"})
		assert.Error(t, err)
	})

	t.Run("Stop", func(t *testing.T) {
		_, err := plugin.Stop()
		assert.NoError(t, err)

		// Wait a bit for processes to stop
		time.Sleep(2 * time.Second)
	})
}

// testScreenshotMethods tests all screenshot methods
func testScreenshotMethods(t *testing.T, plugin computeruse.IComputerUse) {
	t.Run("TakeScreenshot", func(t *testing.T) {
		req := &computeruse.ScreenshotRequest{
			ShowCursor: false,
		}

		resp, err := plugin.TakeScreenshot(req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.Screenshot)

		// Test with cursor
		req.ShowCursor = true
		resp, err = plugin.TakeScreenshot(req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.Screenshot)
		assert.NotNil(t, resp.CursorPosition)
	})

	t.Run("TakeRegionScreenshot", func(t *testing.T) {
		req := &computeruse.RegionScreenshotRequest{
			Position: computeruse.Position{
				X: 50,
				Y: 50,
			},
			Size: computeruse.Size{
				Width:  100,
				Height: 100,
			},
			ShowCursor: false,
		}

		resp, err := plugin.TakeRegionScreenshot(req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.Screenshot)
	})

	t.Run("TakeCompressedScreenshot", func(t *testing.T) {
		req := &computeruse.CompressedScreenshotRequest{
			ShowCursor: false,
			Format:     "png",
			Quality:    85,
			Scale:      0.5,
		}

		resp, err := plugin.TakeCompressedScreenshot(req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.Screenshot)
		assert.True(t, resp.SizeBytes > 0)

		// Test JPEG format
		req.Format = "jpeg"
		req.Quality = 75
		resp, err = plugin.TakeCompressedScreenshot(req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("TakeCompressedRegionScreenshot", func(t *testing.T) {
		req := &computeruse.CompressedRegionScreenshotRequest{
			Position: computeruse.Position{
				X: 50,
				Y: 50,
			},
			Size: computeruse.Size{
				Width:  100,
				Height: 100,
			},
			ShowCursor: false,
			Format:     "png",
			Quality:    90,
			Scale:      0.8,
		}

		resp, err := plugin.TakeCompressedRegionScreenshot(req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.Screenshot)
		assert.True(t, resp.SizeBytes > 0)
	})
}

// testMouseControlMethods tests all mouse control methods
func testMouseControlMethods(t *testing.T, plugin computeruse.IComputerUse) {
	t.Run("GetMousePosition", func(t *testing.T) {
		resp, err := plugin.GetMousePosition()
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.X >= 0)
		assert.True(t, resp.Y >= 0)
	})

	t.Run("MoveMouse", func(t *testing.T) {
		req := &computeruse.MoveMouseRequest{
			Position: computeruse.Position{
				X: 500,
				Y: 300,
			},
		}

		resp, err := plugin.MoveMouse(req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.X >= 0)
		assert.True(t, resp.Y >= 0)
	})

	t.Run("Click", func(t *testing.T) {
		req := &computeruse.ClickRequest{
			Position: computeruse.Position{
				X: 400,
				Y: 200,
			},
			Button: "left",
			Double: false,
		}

		resp, err := plugin.Click(req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, req.X, resp.X)
		assert.Equal(t, req.Y, resp.Y)

		// Test double click
		req.Double = true
		resp, err = plugin.Click(req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Test right click
		req.Button = "right"
		req.Double = false
		resp, err = plugin.Click(req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("Drag", func(t *testing.T) {
		req := &computeruse.DragRequest{
			StartX: 100,
			StartY: 100,
			EndX:   200,
			EndY:   200,
			Button: "left",
		}

		resp, err := plugin.Drag(req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.True(t, resp.X >= 0)
		assert.True(t, resp.Y >= 0)
	})

	t.Run("Scroll", func(t *testing.T) {
		req := &computeruse.ScrollRequest{
			Position: computeruse.Position{
				X: 300,
				Y: 300,
			},
			Direction: "up",
			Amount:    3,
		}

		_, err := plugin.Scroll(req)
		assert.NoError(t, err)

		// Test scroll down
		req.Direction = "down"
		_, err = plugin.Scroll(req)
		assert.NoError(t, err)
	})
}

// testKeyboardControlMethods tests all keyboard control methods
func testKeyboardControlMethods(t *testing.T, plugin computeruse.IComputerUse) {
	t.Run("TypeText", func(t *testing.T) {
		req := &computeruse.TypeTextRequest{
			Text:  "Hello, World!",
			Delay: 10,
		}

		_, err := plugin.TypeText(req)
		assert.NoError(t, err)

		// Test without delay
		req.Delay = 0
		_, err = plugin.TypeText(req)
		assert.NoError(t, err)
	})

	t.Run("PressKey", func(t *testing.T) {
		req := &computeruse.PressKeyRequest{
			Key:       "a",
			Modifiers: []string{},
		}

		_, err := plugin.PressKey(req)
		assert.NoError(t, err)

		// Test with modifiers
		req.Modifiers = []string{"ctrl"}
		_, err = plugin.PressKey(req)
		assert.NoError(t, err)
	})

	t.Run("PressHotkey", func(t *testing.T) {
		req := &computeruse.PressHotkeyRequest{
			Keys: "ctrl+c",
		}

		_, err := plugin.PressHotkey(req)
		assert.NoError(t, err)

		// Test different hotkey
		req.Keys = "alt+tab"
		_, err = plugin.PressHotkey(req)
		assert.NoError(t, err)
	})
}

// testDisplayInfoMethods tests all display info methods
func testDisplayInfoMethods(t *testing.T, plugin computeruse.IComputerUse) {
	t.Run("GetDisplayInfo", func(t *testing.T) {
		resp, err := plugin.GetDisplayInfo()
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Displays)
		assert.True(t, len(resp.Displays) > 0)

		// Check first display
		display := resp.Displays[0]
		assert.True(t, display.ID >= 0)
		assert.True(t, display.Width > 0)
		assert.True(t, display.Height > 0)
		assert.True(t, display.X >= 0)
		assert.True(t, display.Y >= 0)
	})

	t.Run("GetWindows", func(t *testing.T) {
		resp, err := plugin.GetWindows()
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Windows)
		// Note: In test environment, there might not be any windows
		// so we just check that the method doesn't error
	})
}

// testStatusMethod tests the status method
func testStatusMethod(t *testing.T, plugin computeruse.IComputerUse) {
	resp, err := plugin.GetStatus()
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "ok", resp.Status)
}

// TestPluginIntegration tests the plugin as a complete system
func TestPluginIntegration(t *testing.T) {
	// Skip if running in CI or headless environment
	if os.Getenv("CI") == "true" || os.Getenv("DISPLAY") == "" {
		t.Skip("Skipping GUI tests in CI or headless environment")
	}

	client, err := NewTestClient()
	require.NoError(t, err)
	defer client.Close()

	// Test complete workflow
	t.Run("CompleteWorkflow", func(t *testing.T) {
		// 1. Initialize
		_, err := client.plugin.Initialize()
		assert.NoError(t, err)

		// 2. Start processes
		_, err = client.plugin.Start()
		assert.NoError(t, err)
		time.Sleep(3 * time.Second)

		// 3. Check status
		status, err := client.plugin.GetProcessStatus()
		assert.NoError(t, err)
		assert.NotNil(t, status)

		// 4. Take a screenshot
		screenshot, err := client.plugin.TakeScreenshot(&computeruse.ScreenshotRequest{ShowCursor: true})
		assert.NoError(t, err)
		assert.NotNil(t, screenshot)

		// 5. Move mouse and click
		_, err = client.plugin.MoveMouse(&computeruse.MoveMouseRequest{Position: computeruse.Position{X: 100, Y: 100}})
		assert.NoError(t, err)

		_, err = client.plugin.Click(&computeruse.ClickRequest{Position: computeruse.Position{X: 200, Y: 200}, Button: "left"})
		assert.NoError(t, err)

		// 6. Type some text
		_, err = client.plugin.TypeText(&computeruse.TypeTextRequest{Text: "Test"})
		assert.NoError(t, err)

		// 7. Get display info
		displayInfo, err := client.plugin.GetDisplayInfo()
		assert.NoError(t, err)
		assert.NotNil(t, displayInfo)

		// 8. Stop processes
		_, err = client.plugin.Stop()
		assert.NoError(t, err)
		time.Sleep(2 * time.Second)
	})
}

// TestErrorHandling tests error conditions
func TestErrorHandling(t *testing.T) {
	// Skip if running in CI or headless environment
	if os.Getenv("CI") == "true" || os.Getenv("DISPLAY") == "" {
		t.Skip("Skipping GUI tests in CI or headless environment")
	}

	client, err := NewTestClient()
	require.NoError(t, err)
	defer client.Close()

	// Initialize the plugin
	_, err = client.plugin.Initialize()
	require.NoError(t, err)

	t.Run("InvalidProcessName", func(t *testing.T) {
		// Test with non-existing process
		_, err := client.plugin.IsProcessRunning(&computeruse.ProcessRequest{ProcessName: "nonexistent"})
		assert.Error(t, err)

		_, err = client.plugin.RestartProcess(&computeruse.ProcessRequest{ProcessName: "nonexistent"})
		assert.Error(t, err)

		_, err = client.plugin.GetProcessLogs(&computeruse.ProcessRequest{ProcessName: "nonexistent"})
		assert.Error(t, err)

		_, err = client.plugin.GetProcessErrors(&computeruse.ProcessRequest{ProcessName: "nonexistent"})
		assert.Error(t, err)
	})

	t.Run("InvalidHotkey", func(t *testing.T) {
		// Test invalid hotkey format
		_, err := client.plugin.PressHotkey(&computeruse.PressHotkeyRequest{Keys: "invalid"})
		assert.Error(t, err)
	})

	t.Run("InvalidRegionScreenshot", func(t *testing.T) {
		// Test with invalid region (negative coordinates)
		req := &computeruse.RegionScreenshotRequest{
			Position: computeruse.Position{
				X: -100,
				Y: -100,
			},
			Size: computeruse.Size{
				Width:  100,
				Height: 100,
			},
		}

		// This might not error in all environments, but should handle gracefully
		resp, err := client.plugin.TakeRegionScreenshot(req)
		if err != nil {
			t.Logf("TakeRegionScreenshot with negative coordinates returned error (expected): %v", err)
		} else {
			assert.NotNil(t, resp)
		}
	})
}

// TestConcurrentAccess tests concurrent access to the plugin
func TestConcurrentAccess(t *testing.T) {
	// Skip if running in CI or headless environment
	if os.Getenv("CI") == "true" || os.Getenv("DISPLAY") == "" {
		t.Skip("Skipping GUI tests in CI or headless environment")
	}

	client, err := NewTestClient()
	require.NoError(t, err)
	defer client.Close()

	// Initialize the plugin
	_, err = client.plugin.Initialize()
	require.NoError(t, err)

	t.Run("ConcurrentScreenshots", func(t *testing.T) {
		const numGoroutines = 5
		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer func() { done <- true }()

				req := &computeruse.ScreenshotRequest{ShowCursor: false}
				resp, err := client.plugin.TakeScreenshot(req)
				if err != nil {
					t.Errorf("Goroutine %d: TakeScreenshot failed: %v", id, err)
					return
				}
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.Screenshot)
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}
	})

	t.Run("ConcurrentMouseOperations", func(t *testing.T) {
		const numGoroutines = 3
		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer func() { done <- true }()

				// Move mouse to different positions
				x := 100 + (id * 50)
				y := 100 + (id * 50)

				_, err := client.plugin.MoveMouse(&computeruse.MoveMouseRequest{Position: computeruse.Position{X: x, Y: y}})
				if err != nil {
					t.Errorf("Goroutine %d: MoveMouse failed: %v", id, err)
					return
				}
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}
	})
}

// TestPerformance tests performance of key operations
func TestPerformance(t *testing.T) {
	// Skip if running in CI or headless environment
	if os.Getenv("CI") == "true" || os.Getenv("DISPLAY") == "" {
		t.Skip("Skipping GUI tests in CI or headless environment")
	}

	client, err := NewTestClient()
	require.NoError(t, err)
	defer client.Close()

	// Initialize the plugin
	_, err = client.plugin.Initialize()
	require.NoError(t, err)

	t.Run("ScreenshotPerformance", func(t *testing.T) {
		const numScreenshots = 10
		start := time.Now()

		for i := 0; i < numScreenshots; i++ {
			_, err := client.plugin.TakeScreenshot(&computeruse.ScreenshotRequest{ShowCursor: false})
			assert.NoError(t, err)
		}

		duration := time.Since(start)
		avgTime := duration / numScreenshots

		t.Logf("Average screenshot time: %v", avgTime)
		assert.True(t, avgTime < 2*time.Second, "Screenshot should complete within 2 seconds on average")
	})

	t.Run("MouseOperationPerformance", func(t *testing.T) {
		const numOperations = 20
		start := time.Now()

		for i := 0; i < numOperations; i++ {
			x := 100 + (i * 10)
			y := 100 + (i * 10)

			_, err := client.plugin.MoveMouse(&computeruse.MoveMouseRequest{Position: computeruse.Position{X: x, Y: y}})
			assert.NoError(t, err)
		}

		duration := time.Since(start)
		avgTime := duration / numOperations

		t.Logf("Average mouse operation time: %v", avgTime)
		assert.True(t, avgTime < 100*time.Millisecond, "Mouse operation should complete within 100ms on average")
	})
}
