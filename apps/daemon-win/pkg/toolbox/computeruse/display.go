// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"github.com/kbinani/screenshot"
)

// GetDisplayInfo returns information about all available displays
func (c *ComputerUse) GetDisplayInfo() (*DisplayInfoResponse, error) {
	n := screenshot.NumActiveDisplays()
	displays := make([]DisplayInfo, n)

	for i := 0; i < n; i++ {
		bounds := screenshot.GetDisplayBounds(i)
		displays[i] = DisplayInfo{
			ID: i,
			Position: Position{
				X: bounds.Min.X,
				Y: bounds.Min.Y,
			},
			Size: Size{
				Width:  bounds.Dx(),
				Height: bounds.Dy(),
			},
			IsActive: true, // Assuming all detected displays are active
		}
	}

	return &DisplayInfoResponse{
		Displays: displays,
	}, nil
}

// GetWindows returns information about all open windows
func (c *ComputerUse) GetWindows() (*WindowsResponse, error) {
	windowsList := getWindowsList()

	windows := make([]WindowInfo, 0, len(windowsList))
	for i, w := range windowsList {
		windows = append(windows, WindowInfo{
			ID:    i,
			Title: w.Title,
			Position: Position{
				X: 0, // Window position requires additional Windows API calls
				Y: 0,
			},
			Size: Size{
				Width:  0, // Window size requires additional Windows API calls
				Height: 0,
			},
			IsActive: w.Visible,
		})
	}

	return &WindowsResponse{
		Windows: windows,
	}, nil
}
