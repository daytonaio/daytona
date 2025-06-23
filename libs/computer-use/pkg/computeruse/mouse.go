// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"net/http"
	"time"

	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/gin-gonic/gin"
	"github.com/go-vgo/robotgo"
)

func (u *ComputerUse) GetMousePosition(req *computeruse.ComputerUseRequest) (*computeruse.Empty, error) {
	x, y := robotgo.Location()

	req.RequestContext.JSON(http.StatusOK, gin.H{
		"x": x,
		"y": y,
	})

	return new(computeruse.Empty), nil
}

func (u *ComputerUse) MoveMouse(req *computeruse.ComputerUseRequest) (*computeruse.Empty, error) {
	var coords struct {
		X int `json:"x"`
		Y int `json:"y"`
	}

	if err := req.RequestContext.ShouldBindJSON(&coords); err != nil {
		req.RequestContext.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid coordinates",
		})
		return new(computeruse.Empty), nil
	}

	robotgo.MoveMouse(coords.X, coords.Y)

	// Small delay to ensure movement completes
	time.Sleep(50 * time.Millisecond)

	// Verify the mouse actually moved
	actualX, actualY := robotgo.GetMousePos()

	req.RequestContext.JSON(http.StatusOK, gin.H{
		"success":  true,
		"x":        coords.X,
		"y":        coords.Y,
		"actual_x": actualX,
		"actual_y": actualY,
	})

	return new(computeruse.Empty), nil
}

func (u *ComputerUse) Click(req *computeruse.ComputerUseRequest) (*computeruse.Empty, error) {
	var click struct {
		X      int    `json:"x"`
		Y      int    `json:"y"`
		Button string `json:"button"` // left, right, middle
		Double bool   `json:"double"`
	}

	if err := req.RequestContext.ShouldBindJSON(&click); err != nil {
		req.RequestContext.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid click parameters",
		})
		return new(computeruse.Empty), nil
	}

	// Default to left button
	if click.Button == "" {
		click.Button = "left"
	}

	// Move mouse to position first
	robotgo.MoveMouse(click.X, click.Y)
	time.Sleep(100 * time.Millisecond) // Wait for mouse to move

	// Perform the click
	if click.Double {
		robotgo.Click(click.Button, true)
	} else {
		robotgo.Click(click.Button, false)
	}

	// Verify position after click
	actualX, actualY := robotgo.GetMousePos()

	req.RequestContext.JSON(http.StatusOK, gin.H{
		"success":  true,
		"action":   "click",
		"button":   click.Button,
		"double":   click.Double,
		"x":        click.X,
		"y":        click.Y,
		"actual_x": actualX,
		"actual_y": actualY,
	})

	return new(computeruse.Empty), nil
}

// Helper function to move mouse smoothly in steps
func moveMouseSmoothly(startX, startY, endX, endY, steps int) {
	dx := float64(endX-startX) / float64(steps)
	dy := float64(endY-startY) / float64(steps)
	for i := 1; i <= steps; i++ {
		x := int(float64(startX) + dx*float64(i))
		y := int(float64(startY) + dy*float64(i))
		robotgo.MoveMouse(x, y)
		time.Sleep(2 * time.Millisecond)
	}
}

func (u *ComputerUse) Drag(req *computeruse.ComputerUseRequest) (*computeruse.Empty, error) {
	var drag struct {
		StartX int    `json:"startX"`
		StartY int    `json:"startY"`
		EndX   int    `json:"endX"`
		EndY   int    `json:"endY"`
		Button string `json:"button"`
	}

	if err := req.RequestContext.ShouldBindJSON(&drag); err != nil {
		req.RequestContext.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid drag parameters",
		})
		return new(computeruse.Empty), nil
	}

	// Default to left button
	if drag.Button == "" {
		drag.Button = "left"
	}

	// Move to start position
	robotgo.MoveMouse(drag.StartX, drag.StartY)
	time.Sleep(100 * time.Millisecond)

	// Click to focus window before drag
	robotgo.Click(drag.Button, false)
	time.Sleep(100 * time.Millisecond)

	// Ensure mouse button is up before starting
	robotgo.MouseUp(drag.Button)
	time.Sleep(50 * time.Millisecond)

	// Press and hold mouse button
	robotgo.MouseDown(drag.Button)
	time.Sleep(300 * time.Millisecond) // Increased delay

	// Move to end position while holding (smoothly)
	moveMouseSmoothly(drag.StartX, drag.StartY, drag.EndX, drag.EndY, 20)
	time.Sleep(100 * time.Millisecond)

	// Release mouse button
	robotgo.MouseUp(drag.Button)
	time.Sleep(50 * time.Millisecond)

	// Verify final position
	actualX, actualY := robotgo.GetMousePos()

	req.RequestContext.JSON(http.StatusOK, gin.H{
		"success":  true,
		"action":   "drag",
		"from":     gin.H{"x": drag.StartX, "y": drag.StartY},
		"to":       gin.H{"x": drag.EndX, "y": drag.EndY},
		"actual_x": actualX,
		"actual_y": actualY,
	})

	return new(computeruse.Empty), nil
}

func (u *ComputerUse) Scroll(req *computeruse.ComputerUseRequest) (*computeruse.Empty, error) {
	var scroll struct {
		X         int    `json:"x"`
		Y         int    `json:"y"`
		Direction string `json:"direction"` // up, down
		Amount    int    `json:"amount"`
	}

	if err := req.RequestContext.ShouldBindJSON(&scroll); err != nil {
		req.RequestContext.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid scroll parameters",
		})
		return new(computeruse.Empty), nil
	}

	// Default amount if not specified
	if scroll.Amount == 0 {
		scroll.Amount = 3
	}

	// Move mouse to scroll position
	robotgo.MoveMouse(scroll.X, scroll.Y)
	time.Sleep(50 * time.Millisecond)

	// Perform scroll
	if scroll.Direction == "up" {
		robotgo.Scroll(0, scroll.Amount)
	} else if scroll.Direction == "down" {
		robotgo.Scroll(0, -scroll.Amount)
	} else {
		req.RequestContext.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid scroll direction. Use 'up' or 'down'",
		})
		return new(computeruse.Empty), nil
	}

	req.RequestContext.JSON(http.StatusOK, gin.H{
		"success":   true,
		"action":    "scroll",
		"direction": scroll.Direction,
		"amount":    scroll.Amount,
		"x":         scroll.X,
		"y":         scroll.Y,
	})

	return new(computeruse.Empty), nil
}
