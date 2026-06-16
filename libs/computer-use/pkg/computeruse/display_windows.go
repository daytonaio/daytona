//go:build windows

// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/kbinani/screenshot"
)

// GetDisplayInfo returns information about all available displays
func (c *ComputerUse) GetDisplayInfo() (*computeruse.DisplayInfoResponse, error) {
	n := screenshot.NumActiveDisplays()
	displays := make([]computeruse.DisplayInfo, n)

	for i := 0; i < n; i++ {
		bounds := screenshot.GetDisplayBounds(i)
		displays[i] = computeruse.DisplayInfo{
			ID: i,
			Position: computeruse.Position{
				X: bounds.Min.X,
				Y: bounds.Min.Y,
			},
			Size: computeruse.Size{
				Width:  bounds.Dx(),
				Height: bounds.Dy(),
			},
			IsActive: true, // Assuming all detected displays are active
		}
	}

	return &computeruse.DisplayInfoResponse{
		Displays: displays,
	}, nil
}

// GetWindows returns information about all open windows
func (c *ComputerUse) GetWindows() (*computeruse.WindowsResponse, error) {
	windowsList, err := getWindowsList()
	if err != nil {
		return nil, err
	}

	windows := make([]computeruse.WindowInfo, 0, len(windowsList))
	for i, w := range windowsList {
		windows = append(windows, computeruse.WindowInfo{
			ID:    i,
			Title: w.Title,
			Position: computeruse.Position{
				X: w.X,
				Y: w.Y,
			},
			Size: computeruse.Size{
				Width:  w.Width,
				Height: w.Height,
			},
			IsActive: w.Visible,
		})
	}

	return &computeruse.WindowsResponse{
		Windows: windows,
	}, nil
}
