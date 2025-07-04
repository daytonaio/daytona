// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"os"
	"time"

	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/go-vgo/robotgo"
	log "github.com/sirupsen/logrus"
)

func (u *ComputerUse) GetMousePosition() (*computeruse.MousePositionResponse, error) {
	// Debug: Check DISPLAY environment variable
	display := os.Getenv("DISPLAY")
	log.Infof("GetMousePosition: DISPLAY=%s", display)

	x, y := robotgo.Location()

	return &computeruse.MousePositionResponse{
		Position: computeruse.Position{
			X: x,
			Y: y,
		},
	}, nil
}

func (u *ComputerUse) MoveMouse(req *computeruse.MoveMouseRequest) (*computeruse.MousePositionResponse, error) {
	robotgo.Move(req.X, req.Y)

	// Small delay to ensure movement completes
	time.Sleep(50 * time.Millisecond)

	// Get the mouse position after move
	actualX, actualY := robotgo.Location()

	return &computeruse.MousePositionResponse{
		Position: computeruse.Position{
			X: actualX,
			Y: actualY,
		},
	}, nil
}

func (u *ComputerUse) Click(req *computeruse.ClickRequest) (*computeruse.MouseClickResponse, error) {
	// Default to left button
	if req.Button == "" {
		req.Button = "left"
	}

	// Move mouse to position first
	robotgo.Move(req.X, req.Y)
	time.Sleep(100 * time.Millisecond) // Wait for mouse to move

	// Perform the click
	if req.Double {
		robotgo.Click(req.Button, true)
	} else {
		robotgo.Click(req.Button, false)
	}

	// Get position after click
	actualX, actualY := robotgo.Location()

	return &computeruse.MouseClickResponse{
		Position: computeruse.Position{
			X: actualX,
			Y: actualY,
		},
	}, nil
}

// Helper function to move mouse smoothly in steps
func moveMouseSmoothly(startX, startY, endX, endY, steps int) {
	dx := float64(endX-startX) / float64(steps)
	dy := float64(endY-startY) / float64(steps)
	for i := 1; i <= steps; i++ {
		x := int(float64(startX) + dx*float64(i))
		y := int(float64(startY) + dy*float64(i))
		robotgo.Move(x, y)
		time.Sleep(2 * time.Millisecond)
	}
}

func (u *ComputerUse) Drag(req *computeruse.DragRequest) (*computeruse.MouseDragResponse, error) {
	// Default to left button
	if req.Button == "" {
		req.Button = "left"
	}

	// Move to start position
	robotgo.Move(req.StartX, req.StartY)
	time.Sleep(100 * time.Millisecond)

	// Click to focus window before drag
	robotgo.Click(req.Button, false)
	time.Sleep(100 * time.Millisecond)

	// Ensure mouse button is up before starting
	err := robotgo.MouseUp(req.Button)
	if err != nil {
		return nil, err
	}
	time.Sleep(50 * time.Millisecond)

	// Press and hold mouse button
	err = robotgo.MouseDown(req.Button)
	if err != nil {
		return nil, err
	}
	time.Sleep(300 * time.Millisecond) // Increased delay

	// Move to end position while holding (smoothly)
	moveMouseSmoothly(req.StartX, req.StartY, req.EndX, req.EndY, 20)
	time.Sleep(100 * time.Millisecond)

	// Release mouse button
	err = robotgo.MouseUp(req.Button)
	if err != nil {
		return nil, err
	}
	time.Sleep(50 * time.Millisecond)

	// Get final position
	actualX, actualY := robotgo.Location()

	return &computeruse.MouseDragResponse{
		Position: computeruse.Position{
			X: actualX,
			Y: actualY,
		},
	}, nil
}

func (u *ComputerUse) Scroll(req *computeruse.ScrollRequest) (*computeruse.ScrollResponse, error) {
	// Default amount if not specified
	if req.Amount == 0 {
		req.Amount = 3
	}

	// Move mouse to scroll position
	robotgo.Move(req.X, req.Y)
	time.Sleep(50 * time.Millisecond)

	// Perform scroll
	if req.Direction == "up" {
		robotgo.ScrollSmooth(req.Amount, 0)
	} else {
		robotgo.ScrollSmooth(-req.Amount, 0)
	}

	return &computeruse.ScrollResponse{
		Success: true,
	}, nil
}
