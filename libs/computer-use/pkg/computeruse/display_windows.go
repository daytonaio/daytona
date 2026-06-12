//go:build windows

// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
)

// GetWindows returns information about all open windows
func (c *ComputerUse) GetWindows() (*computeruse.WindowsResponse, error) {
	windowsList, err := getWindowsList()
	if err != nil {
		return nil, err
	}
	foreground := getForegroundWindow()

	windows := make([]computeruse.WindowInfo, 0, len(windowsList))
	for _, w := range windowsList {
		windows = append(windows, computeruse.WindowInfo{
			// HWNDs are 32-bit significant per Win32 handle guarantees, so
			// the int conversion is lossless even on 64-bit builds.
			ID:    int(w.HWND),
			Title: w.Title,
			Position: computeruse.Position{
				X: w.X,
				Y: w.Y,
			},
			Size: computeruse.Size{
				Width:  w.Width,
				Height: w.Height,
			},
			IsActive: w.HWND == foreground,
		})
	}

	return &computeruse.WindowsResponse{
		Windows: windows,
	}, nil
}
