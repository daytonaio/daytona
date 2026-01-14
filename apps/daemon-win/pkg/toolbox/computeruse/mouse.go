// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"time"
)

// GetMousePosition returns the current mouse cursor position
func (c *ComputerUse) GetMousePosition() (*MousePositionResponse, error) {
	x, y := getMousePosition()

	return &MousePositionResponse{
		Position: Position{
			X: x,
			Y: y,
		},
	}, nil
}

// MoveMouse moves the mouse cursor to the specified coordinates
func (c *ComputerUse) MoveMouse(req *MouseMoveRequest) (*MousePositionResponse, error) {
	if err := setMousePosition(req.X, req.Y); err != nil {
		return nil, err
	}

	// Small delay to ensure movement completes
	time.Sleep(50 * time.Millisecond)

	// Get the mouse position after move
	actualX, actualY := getMousePosition()

	return &MousePositionResponse{
		Position: Position{
			X: actualX,
			Y: actualY,
		},
	}, nil
}

// Click performs a mouse click at the specified coordinates
func (c *ComputerUse) Click(req *MouseClickRequest) (*MouseClickResponse, error) {
	// Default to left button
	if req.Button == "" {
		req.Button = "left"
	}

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
	actualX, actualY := getMousePosition()

	return &MouseClickResponse{
		Position: Position{
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
		setMousePosition(x, y)
		time.Sleep(2 * time.Millisecond)
	}
}

// Drag performs a mouse drag from start to end coordinates
func (c *ComputerUse) Drag(req *MouseDragRequest) (*MouseDragResponse, error) {
	// Default to left button
	if req.Button == "" {
		req.Button = "left"
	}

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
	actualX, actualY := getMousePosition()

	return &MouseDragResponse{
		Position: Position{
			X: actualX,
			Y: actualY,
		},
	}, nil
}

// Scroll scrolls the mouse wheel at the specified coordinates
func (c *ComputerUse) Scroll(req *MouseScrollRequest) (*ScrollResponse, error) {
	// Default amount if not specified
	if req.Amount == 0 {
		req.Amount = 3
	}

	// Move mouse to scroll position
	if err := setMousePosition(req.X, req.Y); err != nil {
		return nil, err
	}
	time.Sleep(50 * time.Millisecond)

	// Perform scroll
	if err := mouseScroll(req.Amount, req.Direction); err != nil {
		return nil, err
	}

	return &ScrollResponse{
		Success: true,
	}, nil
}
