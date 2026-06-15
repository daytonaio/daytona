// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"image"

	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/go-vgo/robotgo"
	"github.com/kbinani/screenshot"
)

func capturePrimaryDisplay() (*image.RGBA, error) {
	return screenshot.CaptureDisplay(0)
}

func captureRect(rect image.Rectangle) (*image.RGBA, error) {
	return screenshot.CaptureRect(rect)
}

func getCursorPosition(showCursor bool, offsetX, offsetY int) (int, int) {
	if !showCursor {
		return 0, 0
	}

	absoluteMouseX, absoluteMouseY := robotgo.Location()
	return absoluteMouseX - offsetX, absoluteMouseY - offsetY
}

func captureWithCursor(showCursor bool, offsetX, offsetY int, capture func() (*image.RGBA, error)) (*image.RGBA, int, int, error) {
	var img *image.RGBA
	mouseX, mouseY := 0, 0

	err := withX11Client(func() error {
		var err error
		img, err = capture()
		if err != nil {
			return err
		}
		mouseX, mouseY = getCursorPosition(showCursor, offsetX, offsetY)
		return nil
	})
	if err != nil {
		return nil, 0, 0, err
	}

	return img, mouseX, mouseY, nil
}

func getDisplayInfos() ([]computeruse.DisplayInfo, error) {
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
			IsActive: true, // Assuming all detected displays are active.
		}
	}

	return displays, nil
}
