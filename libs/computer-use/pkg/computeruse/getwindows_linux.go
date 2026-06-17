// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build linux

package computeruse

import (
	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/robotn/xgbutil"
	"github.com/robotn/xgbutil/ewmh"
)

func getWindows() ([]computeruse.WindowInfo, error) {
	return getWindowsX11()
}

func getWindowsX11() ([]computeruse.WindowInfo, error) {
	xu, err := xgbutil.NewConn()
	if err != nil {
		return nil, err
	}
	defer xu.Conn().Close()

	clientList, err := ewmh.ClientListGet(xu)
	if err != nil {
		return nil, err
	}

	windows := make([]computeruse.WindowInfo, 0, len(clientList))
	for _, win := range clientList {
		title, err := ewmh.WmVisibleNameGet(xu, win)
		if err != nil || title == "" {
			title, _ = ewmh.WmNameGet(xu, win)
		}
		if title == "" {
			continue
		}

		id := 0
		if pid, err := ewmh.WmPidGet(xu, win); err == nil {
			id = int(pid)
		}

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

	return windows, nil
}
