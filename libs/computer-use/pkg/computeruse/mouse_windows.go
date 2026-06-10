//go:build windows

// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"time"
)

// GetMousePosition returns the current mouse cursor position
func (c *ComputerUse) GetMousePosition() (*computeruse.MousePositionResponse, error) {
	x, y, err := getMousePositionChecked()
	if err != nil {
		return nil, err
	}

	return &computeruse.MousePositionResponse{
		Position: computeruse.Position{
			X: x,
			Y: y,
		},
	}, nil
}

// MoveMouse moves the mouse cursor to the specified coordinates
func (c *ComputerUse) MoveMouse(req *computeruse.MouseMoveRequest) (*computeruse.MousePositionResponse, error) {
	if err := setMousePosition(req.X, req.Y); err != nil {
		return nil, err
	}

	// Small delay to ensure movement completes
	time.Sleep(50 * time.Millisecond)

	// Get the mouse position after move
	actualX, actualY, err := getMousePositionChecked()
	if err != nil {
		return nil, err
	}

	return &computeruse.MousePositionResponse{
		Position: computeruse.Position{
			X: actualX,
			Y: actualY,
		},
	}, nil
}

// Click performs a mouse click at the specified coordinates
func (c *ComputerUse) Click(req *computeruse.MouseClickRequest) (*computeruse.MouseClickResponse, error) {
	button, err := normalizeMouseButton(req.Button)
	if err != nil {
		return nil, err
	}
	req.Button = button

	// Move mouse to position first
	if err := setMousePosition(req.X, req.Y); err != nil {
		return nil, err
	}
	time.Sleep(100 * time.Millisecond) // Wait for mouse to move

	// Perform the click
	if err := mouseClick(req.Button, req.Double); err != nil {
		return nil, err
	}

	// Get position after click
	actualX, actualY, err := getMousePositionChecked()
	if err != nil {
		return nil, err
	}

	return &computeruse.MouseClickResponse{
		Position: computeruse.Position{
			X: actualX,
			Y: actualY,
		},
	}, nil
}

// moveMouseSmoothly moves the mouse smoothly in steps
func moveMouseSmoothly(startX, startY, endX, endY, steps int) {
	dx := float64(endX-startX) / float64(steps)
	dy := float64(endY-startY) / float64(steps)
	for i := 1; i <= steps; i++ {
		x := int(float64(startX) + dx*float64(i))
		y := int(float64(startY) + dy*float64(i))
		_ = setMousePosition(x, y)
		time.Sleep(2 * time.Millisecond)
	}
}

// Drag performs a mouse drag from start to end coordinates
func (c *ComputerUse) Drag(req *computeruse.MouseDragRequest) (*computeruse.MouseDragResponse, error) {
	button, err := normalizeMouseButton(req.Button)
	if err != nil {
		return nil, err
	}
	req.Button = button

	// Move to start position
	if err := setMousePosition(req.StartX, req.StartY); err != nil {
		return nil, err
	}
	time.Sleep(100 * time.Millisecond)

	// Click to focus window before drag
	if err := mouseClick(req.Button, false); err != nil {
		return nil, err
	}
	time.Sleep(100 * time.Millisecond)

	// Ensure mouse button is up before starting
	if err := mouseUp(req.Button); err != nil {
		return nil, err
	}
	time.Sleep(50 * time.Millisecond)

	// Press and hold mouse button
	if err := mouseDown(req.Button); err != nil {
		return nil, err
	}
	time.Sleep(300 * time.Millisecond) // Increased delay

	// Move to end position while holding (smoothly)
	moveMouseSmoothly(req.StartX, req.StartY, req.EndX, req.EndY, 20)
	time.Sleep(100 * time.Millisecond)

	// Release mouse button
	if err := mouseUp(req.Button); err != nil {
		return nil, err
	}
	time.Sleep(50 * time.Millisecond)

	// Get final position
	actualX, actualY, err := getMousePositionChecked()
	if err != nil {
		return nil, err
	}

	return &computeruse.MouseDragResponse{
		Position: computeruse.Position{
			X: actualX,
			Y: actualY,
		},
	}, nil
}

// Scroll scrolls the mouse wheel at the specified coordinates
func (c *ComputerUse) Scroll(req *computeruse.MouseScrollRequest) (*computeruse.ScrollResponse, error) {
	direction, err := normalizeScrollDirection(req.Direction)
	if err != nil {
		return nil, err
	}
	req.Direction = direction

	amount, err := normalizeScrollAmount(req.Amount)
	if err != nil {
		return nil, err
	}
	req.Amount = amount

	// Move mouse to scroll position
	if err := setMousePosition(req.X, req.Y); err != nil {
		return nil, err
	}
	time.Sleep(50 * time.Millisecond)

	// Perform scroll
	if err := mouseScroll(req.Amount, req.Direction); err != nil {
		return nil, err
	}

	return &computeruse.ScrollResponse{
		Success: true,
	}, nil
}
