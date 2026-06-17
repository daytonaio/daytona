// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build !linux

package computeruse

import (
	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/go-vgo/robotgo"
)

func getWindows() ([]computeruse.WindowInfo, error) {
	// This is a simplified version - robotgo's window functions
	// might need additional setup depending on the platform.

	windows := make([]computeruse.WindowInfo, 0)
	titles, err := robotgo.FindIds("")
	if err != nil {
		return nil, err
	}

	for _, id := range titles {
		title := robotgo.GetTitle(id)
		if title != "" {
			// Get window position and size (this might need platform-specific implementation)
			// For now, we'll use placeholder values.
			windows = append(windows, computeruse.WindowInfo{
				ID:    id,
				Title: title,
				Position: computeruse.Position{
					X: 0, // Would need platform-specific implementation.
					Y: 0, // Would need platform-specific implementation.
				},
				Size: computeruse.Size{
					Width:  0, // Would need platform-specific implementation.
					Height: 0, // Would need platform-specific implementation.
				},
				IsActive: false, // Would need platform-specific implementation.
			})
		}
	}

	return windows, nil
}
