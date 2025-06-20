// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-vgo/robotgo"
)

func (u *ComputerUse) GetMousePosition(c *gin.Context) {
	x, y := robotgo.Location()

	c.JSON(http.StatusOK, gin.H{
		"x": x,
		"y": y,
	})
}

func (u *ComputerUse) MoveMouse(c *gin.Context) {
	var coords struct {
		X int `json:"x"`
		Y int `json:"y"`
	}

	if err := c.ShouldBindJSON(&coords); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid coordinates",
		})
		return
	}

	robotgo.MoveMouse(coords.X, coords.Y)

	// Small delay to ensure movement completes
	time.Sleep(50 * time.Millisecond)

	// Verify the mouse actually moved
	actualX, actualY := robotgo.GetMousePos()

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"x":        coords.X,
		"y":        coords.Y,
		"actual_x": actualX,
		"actual_y": actualY,
	})
}

func (u *ComputerUse) Click(c *gin.Context) {
	var click struct {
		X      int    `json:"x"`
		Y      int    `json:"y"`
		Button string `json:"button"` // left, right, middle
		Double bool   `json:"double"`
	}

	if err := c.ShouldBindJSON(&click); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid click parameters",
		})
		return
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

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"action":   "click",
		"button":   click.Button,
		"double":   click.Double,
		"x":        click.X,
		"y":        click.Y,
		"actual_x": actualX,
		"actual_y": actualY,
	})
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

func (u *ComputerUse) Drag(c *gin.Context) {
	var drag struct {
		StartX int    `json:"startX"`
		StartY int    `json:"startY"`
		EndX   int    `json:"endX"`
		EndY   int    `json:"endY"`
		Button string `json:"button"`
	}

	if err := c.ShouldBindJSON(&drag); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid drag parameters",
		})
		return
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

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"action":   "drag",
		"from":     gin.H{"x": drag.StartX, "y": drag.StartY},
		"to":       gin.H{"x": drag.EndX, "y": drag.EndY},
		"actual_x": actualX,
		"actual_y": actualY,
	})
}

func (u *ComputerUse) Scroll(c *gin.Context) {
	var scroll struct {
		X         int    `json:"x"`
		Y         int    `json:"y"`
		Direction string `json:"direction"` // up, down
		Amount    int    `json:"amount"`
	}

	if err := c.ShouldBindJSON(&scroll); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid scroll parameters",
		})
		return
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid scroll direction. Use 'up' or 'down'",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"action":    "scroll",
		"direction": scroll.Direction,
		"amount":    scroll.Amount,
		"x":         scroll.X,
		"y":         scroll.Y,
	})
}
