// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"

	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/go-vgo/robotgo"
	"github.com/kbinani/screenshot"
	log "github.com/sirupsen/logrus"
)

// drawCursor draws a simple cursor at the given position
func drawCursor(img *image.RGBA, x, y int) {
	// Define cursor colors
	white := color.RGBA{255, 255, 255, 255}
	black := color.RGBA{0, 0, 0, 255}

	// Draw a simple crosshair cursor
	cursorSize := 20

	// Draw white outline (thicker)
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			// Horizontal line
			for dx := -cursorSize; dx <= cursorSize; dx++ {
				px, py := x+dx, y+i
				if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
					img.Set(px, py, white)
				}
			}
			// Vertical line
			for dy := -cursorSize; dy <= cursorSize; dy++ {
				px, py := x+j, y+dy
				if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
					img.Set(px, py, white)
				}
			}
		}
	}

	// Draw black center
	// Horizontal line
	for dx := -cursorSize; dx <= cursorSize; dx++ {
		px := x + dx
		if px >= 0 && px < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
			img.Set(px, y, black)
		}
	}
	// Vertical line
	for dy := -cursorSize; dy <= cursorSize; dy++ {
		py := y + dy
		if x >= 0 && x < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
			img.Set(x, py, black)
		}
	}

	// Draw center dot
	for i := -2; i <= 2; i++ {
		for j := -2; j <= 2; j++ {
			px, py := x+i, y+j
			if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
				img.Set(px, py, color.RGBA{255, 0, 0, 255}) // Red center
			}
		}
	}
}

func (u *ComputerUse) TakeScreenshot(req *computeruse.ScreenshotRequest) (*computeruse.ScreenshotResponse, error) {
	// Debug: Check DISPLAY environment variable
	display := os.Getenv("DISPLAY")
	log.Infof("TakeScreenshot: DISPLAY=%s", display)

	bounds := screenshot.GetDisplayBounds(0)
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		log.Errorf("TakeScreenshot error: %v", err)
		return nil, err
	}

	// Convert to RGBA for drawing
	rgbaImg := image.NewRGBA(img.Bounds())
	draw.Draw(rgbaImg, rgbaImg.Bounds(), img, image.Point{}, draw.Src)

	// Draw cursor if requested
	mouseX, mouseY := 0, 0
	if req.ShowCursor {
		mouseX, mouseY = robotgo.Location()
		drawCursor(rgbaImg, mouseX, mouseY)
	}

	// Convert to base64
	var buf bytes.Buffer
	if err := png.Encode(&buf, rgbaImg); err != nil {
		return nil, err
	}

	base64Str := base64.StdEncoding.EncodeToString(buf.Bytes())

	response := &computeruse.ScreenshotResponse{
		Screenshot: base64Str,
		CursorPosition: &computeruse.Position{
			X: mouseX,
			Y: mouseY,
		},
	}

	if req.ShowCursor {
		response.CursorPosition = &computeruse.Position{
			X: mouseX,
			Y: mouseY,
		}
	}

	return response, nil
}

func (u *ComputerUse) TakeRegionScreenshot(req *computeruse.RegionScreenshotRequest) (*computeruse.ScreenshotResponse, error) {
	// Debug: Check DISPLAY environment variable
	display := os.Getenv("DISPLAY")
	log.Infof("TakeRegionScreenshot: DISPLAY=%s", display)

	rect := image.Rect(req.X, req.Y, req.X+req.Width, req.Y+req.Height)
	img, err := screenshot.CaptureRect(rect)
	if err != nil {
		log.Errorf("TakeRegionScreenshot error: %v", err)
		return nil, err
	}

	// Convert to RGBA for drawing
	rgbaImg := image.NewRGBA(img.Bounds())
	draw.Draw(rgbaImg, rgbaImg.Bounds(), img, image.Point{}, draw.Src)

	// Draw cursor if requested and it's within the region
	mouseX, mouseY := 0, 0
	if req.ShowCursor {
		absoluteMouseX, absoluteMouseY := robotgo.Location()
		// Convert to relative coordinates within the region
		mouseX = absoluteMouseX - req.X
		mouseY = absoluteMouseY - req.Y

		// Only draw if cursor is within the region
		if mouseX >= 0 && mouseX < req.Width && mouseY >= 0 && mouseY < req.Height {
			drawCursor(rgbaImg, mouseX, mouseY)
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, rgbaImg); err != nil {
		return nil, err
	}

	base64Str := base64.StdEncoding.EncodeToString(buf.Bytes())

	response := &computeruse.ScreenshotResponse{
		Screenshot: base64Str,
		CursorPosition: &computeruse.Position{
			X: mouseX + req.X,
			Y: mouseY + req.Y,
		},
	}

	if req.ShowCursor {
		response.CursorPosition = &computeruse.Position{
			X: mouseX + req.X,
			Y: mouseY + req.Y,
		}
	}

	return response, nil
}
