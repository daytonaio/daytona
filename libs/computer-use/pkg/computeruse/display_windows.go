//go:build windows

// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
)

// GetWindows returns information about all open windows
func (c *ComputerUse) GetWindows() (*computeruse.WindowsResponse, error) {
	windowsList := getWindowsList()

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
