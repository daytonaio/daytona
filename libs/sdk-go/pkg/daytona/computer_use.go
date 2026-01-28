// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	"context"

	"github.com/daytonaio/daytona/libs/sdk-go/pkg/errors"
	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
	"github.com/daytonaio/daytona/libs/toolbox-api-client-go"
)

// ComputerUseService provides desktop automation operations for a sandbox.
//
// ComputerUseService enables GUI automation including mouse control, keyboard input,
// screenshots, and display management. The desktop environment must be started
// before using these features. Access through [Sandbox.ComputerUse].
//
// Example:
//
//	cu := sandbox.ComputerUse
//
//	// Start the desktop environment
//	if err := cu.Start(ctx); err != nil {
//	    return err
//	}
//	defer cu.Stop(ctx)
//
//	// Take a screenshot
//	screenshot, err := cu.Screenshot().TakeFullScreen(ctx, nil)
//	if err != nil {
//	    return err
//	}
//
//	// Click at coordinates
//	cu.Mouse().Click(ctx, 100, 200, nil, nil)
//
//	// Type text
//	cu.Keyboard().Type(ctx, "Hello, World!", nil)
type ComputerUseService struct {
	toolboxClient *toolbox.APIClient

	mouse      *MouseService
	keyboard   *KeyboardService
	screenshot *ScreenshotService
	display    *DisplayService
}

// NewComputerUseService creates a new ComputerUseService.
//
// This is typically called internally by the SDK when creating a [Sandbox].
// Users should access ComputerUseService through [Sandbox.ComputerUse] rather than
// creating it directly.
func NewComputerUseService(toolboxClient *toolbox.APIClient) *ComputerUseService {
	return &ComputerUseService{
		toolboxClient: toolboxClient,
	}
}

// Mouse returns the [MouseService] for mouse operations.
//
// The service is lazily initialized on first access.
func (c *ComputerUseService) Mouse() *MouseService {
	if c.mouse == nil {
		c.mouse = NewMouseService(c.toolboxClient)
	}
	return c.mouse
}

// Keyboard returns the [KeyboardService] for keyboard operations.
//
// The service is lazily initialized on first access.
func (c *ComputerUseService) Keyboard() *KeyboardService {
	if c.keyboard == nil {
		c.keyboard = NewKeyboardService(c.toolboxClient)
	}
	return c.keyboard
}

// Screenshot returns the [ScreenshotService] for capturing screen images.
//
// The service is lazily initialized on first access.
func (c *ComputerUseService) Screenshot() *ScreenshotService {
	if c.screenshot == nil {
		c.screenshot = NewScreenshotService(c.toolboxClient)
	}
	return c.screenshot
}

// Display returns the [DisplayService] for display information.
//
// The service is lazily initialized on first access.
func (c *ComputerUseService) Display() *DisplayService {
	if c.display == nil {
		c.display = NewDisplayService(c.toolboxClient)
	}
	return c.display
}

// Start initializes and starts the desktop environment.
//
// The desktop environment must be started before using mouse, keyboard, or
// screenshot operations. Call [ComputerUseService.Stop] when finished.
//
// Example:
//
//	if err := cu.Start(ctx); err != nil {
//	    return err
//	}
//	defer cu.Stop(ctx)
//
// Returns an error if the desktop fails to start.
func (c *ComputerUseService) Start(ctx context.Context) error {
	_, httpResp, err := c.toolboxClient.ComputerUseAPI.StartComputerUse(ctx).Execute()
	if err != nil {
		return errors.ConvertToolboxError(err, httpResp)
	}

	return nil
}

// Stop shuts down the desktop environment and releases resources.
//
// Example:
//
//	err := cu.Stop(ctx)
//
// Returns an error if the desktop fails to stop gracefully.
func (c *ComputerUseService) Stop(ctx context.Context) error {
	_, httpResp, err := c.toolboxClient.ComputerUseAPI.StopComputerUse(ctx).Execute()
	if err != nil {
		return errors.ConvertToolboxError(err, httpResp)
	}

	return nil
}

// GetStatus returns the current status of the desktop environment.
//
// Example:
//
//	status, err := cu.GetStatus(ctx)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Desktop status: %v\n", status["status"])
//
// Returns a map containing status information.
func (c *ComputerUseService) GetStatus(ctx context.Context) (map[string]any, error) {
	status, httpResp, err := c.toolboxClient.ComputerUseAPI.GetComputerUseStatus(ctx).Execute()
	if err != nil {
		return nil, errors.ConvertToolboxError(err, httpResp)
	}

	// Convert to map for backward compatibility
	result := make(map[string]any)
	if status.Status != nil {
		result["status"] = status.GetStatus()
	}

	return result, nil
}

// MouseService provides mouse control operations.
//
// MouseService enables cursor movement, clicking, dragging, and scrolling.
// Access through [ComputerUseService.Mouse].
type MouseService struct {
	toolboxClient *toolbox.APIClient
}

// NewMouseService creates a new MouseService.
func NewMouseService(toolboxClient *toolbox.APIClient) *MouseService {
	return &MouseService{
		toolboxClient: toolboxClient,
	}
}

// GetPosition returns the current cursor position.
//
// Example:
//
//	pos, err := mouse.GetPosition(ctx)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Cursor at (%v, %v)\n", pos["x"], pos["y"])
//
// Returns a map with "x" and "y" coordinates.
func (m *MouseService) GetPosition(ctx context.Context) (map[string]any, error) {
	pos, httpResp, err := m.toolboxClient.ComputerUseAPI.GetMousePosition(ctx).Execute()
	if err != nil {
		return nil, errors.ConvertToolboxError(err, httpResp)
	}

	// Convert to map for backward compatibility
	return map[string]any{
		"x": pos.GetX(),
		"y": pos.GetY(),
	}, nil
}

// Move moves the cursor to the specified coordinates.
//
// Parameters:
//   - x: Target X coordinate
//   - y: Target Y coordinate
//
// Example:
//
//	pos, err := mouse.Move(ctx, 500, 300)
//
// Returns a map with the new "x" and "y" coordinates.
func (m *MouseService) Move(ctx context.Context, x, y int) (map[string]any, error) {
	req := toolbox.NewMouseMoveRequest()
	xInt32 := int32(x)
	yInt32 := int32(y)
	req.SetX(xInt32)
	req.SetY(yInt32)

	pos, httpResp, err := m.toolboxClient.ComputerUseAPI.MoveMouse(ctx).Request(*req).Execute()
	if err != nil {
		return nil, errors.ConvertToolboxError(err, httpResp)
	}

	// Convert to map for backward compatibility
	return map[string]any{
		"x": pos.GetX(),
		"y": pos.GetY(),
	}, nil
}

// Click performs a mouse click at the specified coordinates.
//
// Parameters:
//   - x: X coordinate to click
//   - y: Y coordinate to click
//   - button: Mouse button ("left", "right", "middle"), nil for left click
//   - double: Whether to double-click, nil for single click
//
// Example:
//
//	// Single left click
//	pos, err := mouse.Click(ctx, 100, 200, nil, nil)
//
//	// Right click
//	button := "right"
//	pos, err := mouse.Click(ctx, 100, 200, &button, nil)
//
//	// Double click
//	doubleClick := true
//	pos, err := mouse.Click(ctx, 100, 200, nil, &doubleClick)
//
// Returns a map with the click "x" and "y" coordinates.
func (m *MouseService) Click(ctx context.Context, x, y int, button *string, double *bool) (map[string]any, error) {
	req := toolbox.NewMouseClickRequest()
	xInt32 := int32(x)
	yInt32 := int32(y)
	req.SetX(xInt32)
	req.SetY(yInt32)
	if button != nil {
		req.SetButton(*button)
	}
	if double != nil {
		req.SetDouble(*double)
	}

	result, httpResp, err := m.toolboxClient.ComputerUseAPI.Click(ctx).Request(*req).Execute()
	if err != nil {
		return nil, errors.ConvertToolboxError(err, httpResp)
	}

	// Convert to map for backward compatibility
	return map[string]any{
		"x": result.GetX(),
		"y": result.GetY(),
	}, nil
}

// Drag performs a mouse drag operation from start to end coordinates.
//
// Parameters:
//   - startX, startY: Starting coordinates
//   - endX, endY: Ending coordinates
//   - button: Mouse button to use, nil for left button
//
// Example:
//
//	// Drag from (100, 100) to (300, 300)
//	pos, err := mouse.Drag(ctx, 100, 100, 300, 300, nil)
//
// Returns a map with the final "x" and "y" coordinates.
func (m *MouseService) Drag(ctx context.Context, startX, startY, endX, endY int, button *string) (map[string]any, error) {
	req := toolbox.NewMouseDragRequest()
	req.SetStartX(int32(startX))
	req.SetStartY(int32(startY))
	req.SetEndX(int32(endX))
	req.SetEndY(int32(endY))
	if button != nil {
		req.SetButton(*button)
	}

	result, httpResp, err := m.toolboxClient.ComputerUseAPI.Drag(ctx).Request(*req).Execute()
	if err != nil {
		return nil, errors.ConvertToolboxError(err, httpResp)
	}

	// Convert to map for backward compatibility
	// MouseDragResponse only returns final position
	return map[string]any{
		"x": result.GetX(),
		"y": result.GetY(),
	}, nil
}

// Scroll performs a mouse scroll operation at the specified coordinates.
//
// Parameters:
//   - x, y: Coordinates where the scroll occurs
//   - direction: Scroll direction ("up", "down", "left", "right")
//   - amount: Scroll amount, nil for default
//
// Example:
//
//	// Scroll down at position (500, 400)
//	success, err := mouse.Scroll(ctx, 500, 400, "down", nil)
//
//	// Scroll up with specific amount
//	amount := 5
//	success, err := mouse.Scroll(ctx, 500, 400, "up", &amount)
//
// Returns true if the scroll was successful.
func (m *MouseService) Scroll(ctx context.Context, x, y int, direction string, amount *int) (bool, error) {
	req := toolbox.NewMouseScrollRequest()
	req.SetX(int32(x))
	req.SetY(int32(y))
	req.SetDirection(direction)
	if amount != nil {
		req.SetAmount(int32(*amount))
	}

	result, httpResp, err := m.toolboxClient.ComputerUseAPI.Scroll(ctx).Request(*req).Execute()
	if err != nil {
		return false, errors.ConvertToolboxError(err, httpResp)
	}

	return result.GetSuccess(), nil
}

// KeyboardService provides keyboard input operations.
//
// KeyboardService enables typing text, pressing keys, and executing keyboard shortcuts.
// Access through [ComputerUseService.Keyboard].
type KeyboardService struct {
	toolboxClient *toolbox.APIClient
}

// NewKeyboardService creates a new KeyboardService.
func NewKeyboardService(toolboxClient *toolbox.APIClient) *KeyboardService {
	return &KeyboardService{
		toolboxClient: toolboxClient,
	}
}

// Type simulates typing the specified text.
//
// Parameters:
//   - text: The text to type
//   - delay: Delay in milliseconds between keystrokes, nil for default
//
// Example:
//
//	// Type text with default speed
//	err := keyboard.Type(ctx, "Hello, World!", nil)
//
//	// Type with custom delay between keystrokes
//	delay := 50
//	err := keyboard.Type(ctx, "Slow typing", &delay)
//
// Returns an error if typing fails.
func (k *KeyboardService) Type(ctx context.Context, text string, delay *int) error {
	req := toolbox.NewKeyboardTypeRequest()
	req.SetText(text)
	if delay != nil {
		req.SetDelay(int32(*delay))
	}

	_, httpResp, err := k.toolboxClient.ComputerUseAPI.TypeText(ctx).Request(*req).Execute()
	if err != nil {
		return errors.ConvertToolboxError(err, httpResp)
	}

	return nil
}

// Press simulates pressing a key with optional modifiers.
//
// Parameters:
//   - key: The key to press (e.g., "a", "Enter", "Tab", "F1")
//   - modifiers: Modifier keys to hold (e.g., "ctrl", "alt", "shift", "meta")
//
// Example:
//
//	// Press Enter
//	err := keyboard.Press(ctx, "Enter", nil)
//
//	// Press Ctrl+S
//	err := keyboard.Press(ctx, "s", []string{"ctrl"})
//
//	// Press Ctrl+Shift+N
//	err := keyboard.Press(ctx, "n", []string{"ctrl", "shift"})
//
// Returns an error if the key press fails.
func (k *KeyboardService) Press(ctx context.Context, key string, modifiers []string) error {
	req := toolbox.NewKeyboardPressRequest()
	req.SetKey(key)
	if modifiers != nil {
		req.SetModifiers(modifiers)
	}

	_, httpResp, err := k.toolboxClient.ComputerUseAPI.PressKey(ctx).Request(*req).Execute()
	if err != nil {
		return errors.ConvertToolboxError(err, httpResp)
	}

	return nil
}

// Hotkey executes a keyboard shortcut.
//
// Parameters:
//   - keys: The hotkey combination as a string (e.g., "ctrl+c", "alt+tab")
//
// Example:
//
//	// Copy (Ctrl+C)
//	err := keyboard.Hotkey(ctx, "ctrl+c")
//
//	// Paste (Ctrl+V)
//	err := keyboard.Hotkey(ctx, "ctrl+v")
//
//	// Switch windows (Alt+Tab)
//	err := keyboard.Hotkey(ctx, "alt+tab")
//
// Returns an error if the hotkey fails.
func (k *KeyboardService) Hotkey(ctx context.Context, keys string) error {
	req := toolbox.NewKeyboardHotkeyRequest()
	req.SetKeys(keys)

	_, httpResp, err := k.toolboxClient.ComputerUseAPI.PressHotkey(ctx).Request(*req).Execute()
	if err != nil {
		return errors.ConvertToolboxError(err, httpResp)
	}

	return nil
}

// ScreenshotService provides screen capture operations.
//
// ScreenshotService enables capturing full screen or region screenshots.
// Access through [ComputerUseService.Screenshot].
type ScreenshotService struct {
	toolboxClient *toolbox.APIClient
}

// NewScreenshotService creates a new ScreenshotService.
func NewScreenshotService(toolboxClient *toolbox.APIClient) *ScreenshotService {
	return &ScreenshotService{
		toolboxClient: toolboxClient,
	}
}

// TakeFullScreen captures a screenshot of the entire screen.
//
// Parameters:
//   - showCursor: Whether to include the cursor in the screenshot, nil for default
//
// Example:
//
//	// Capture full screen
//	screenshot, err := ss.TakeFullScreen(ctx, nil)
//	if err != nil {
//	    return err
//	}
//	// screenshot.Image contains the base64-encoded image data
//
//	// Capture with cursor visible
//	showCursor := true
//	screenshot, err := ss.TakeFullScreen(ctx, &showCursor)
//
// Returns [types.ScreenshotResponse] with the captured image.
func (s *ScreenshotService) TakeFullScreen(ctx context.Context, showCursor *bool) (*types.ScreenshotResponse, error) {
	req := s.toolboxClient.ComputerUseAPI.TakeScreenshot(ctx)
	if showCursor != nil {
		req = req.ShowCursor(*showCursor)
	}

	result, httpResp, err := req.Execute()
	if err != nil {
		return nil, errors.ConvertToolboxError(err, httpResp)
	}

	// Note: Toolbox API returns screenshot but not width/height separately
	// Width and height would need to be parsed from the image data if needed
	return &types.ScreenshotResponse{
		Image:     result.GetScreenshot(),
		Width:     0, // Not provided by toolbox API
		Height:    0, // Not provided by toolbox API
		SizeBytes: convertInt32PtrToIntPtr(result.SizeBytes),
	}, nil
}

// TakeRegion captures a screenshot of a specific screen region.
//
// Parameters:
//   - region: The region to capture (X, Y, Width, Height)
//   - showCursor: Whether to include the cursor in the screenshot, nil for default
//
// Example:
//
//	// Capture a 200x100 region starting at (50, 50)
//	region := types.ScreenshotRegion{X: 50, Y: 50, Width: 200, Height: 100}
//	screenshot, err := ss.TakeRegion(ctx, region, nil)
//	if err != nil {
//	    return err
//	}
//
// Returns [types.ScreenshotResponse] with the captured image.
func (s *ScreenshotService) TakeRegion(ctx context.Context, region types.ScreenshotRegion, showCursor *bool) (*types.ScreenshotResponse, error) {
	req := s.toolboxClient.ComputerUseAPI.TakeRegionScreenshot(ctx).
		X(int32(region.X)).
		Y(int32(region.Y)).
		Width(int32(region.Width)).
		Height(int32(region.Height))

	if showCursor != nil {
		req = req.ShowCursor(*showCursor)
	}

	result, httpResp, err := req.Execute()
	if err != nil {
		return nil, errors.ConvertToolboxError(err, httpResp)
	}

	// Note: Toolbox API returns screenshot but not width/height separately
	// The region dimensions are known from the request parameters
	return &types.ScreenshotResponse{
		Image:     result.GetScreenshot(),
		Width:     region.Width,
		Height:    region.Height,
		SizeBytes: convertInt32PtrToIntPtr(result.SizeBytes),
	}, nil
}

// Helper function to convert *int32 to *int
func convertInt32PtrToIntPtr(i *int32) *int {
	if i == nil {
		return nil
	}
	val := int(*i)
	return &val
}

// DisplayService provides display information and window management operations.
//
// DisplayService enables querying display configuration and window information.
// Access through [ComputerUseService.Display].
type DisplayService struct {
	toolboxClient *toolbox.APIClient
}

// NewDisplayService creates a new DisplayService.
func NewDisplayService(toolboxClient *toolbox.APIClient) *DisplayService {
	return &DisplayService{
		toolboxClient: toolboxClient,
	}
}

// GetInfo returns information about connected displays.
//
// Example:
//
//	info, err := display.GetInfo(ctx)
//	if err != nil {
//	    return err
//	}
//	displays := info["displays"]
//	fmt.Printf("Connected displays: %v\n", displays)
//
// Returns a map containing display information.
func (d *DisplayService) GetInfo(ctx context.Context) (map[string]any, error) {
	info, httpResp, err := d.toolboxClient.ComputerUseAPI.GetDisplayInfo(ctx).Execute()
	if err != nil {
		return nil, errors.ConvertToolboxError(err, httpResp)
	}

	// Convert to map for backward compatibility
	return map[string]any{
		"displays": info.GetDisplays(),
	}, nil
}

// GetWindows returns information about open windows.
//
// Example:
//
//	result, err := display.GetWindows(ctx)
//	if err != nil {
//	    return err
//	}
//	windows := result["windows"]
//	fmt.Printf("Open windows: %v\n", windows)
//
// Returns a map containing window information.
func (d *DisplayService) GetWindows(ctx context.Context) (map[string]any, error) {
	windows, httpResp, err := d.toolboxClient.ComputerUseAPI.GetWindows(ctx).Execute()
	if err != nil {
		return nil, errors.ConvertToolboxError(err, httpResp)
	}

	// Convert to map for backward compatibility
	return map[string]any{
		"windows": windows.GetWindows(),
	}, nil
}
