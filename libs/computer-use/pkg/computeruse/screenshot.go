//go:build linux || windows

// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/draw"
	"image/png"

	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/kbinani/screenshot"
)

// The capture pipeline below is shared between Linux and Windows. The only
// platform-specific call is getMousePosition(), implemented per OS (robotgo
// on Linux in computeruse.go, GetCursorPos on Windows in winapi_windows.go).

// TakeScreenshot takes a screenshot of the entire screen
func (c *ComputerUse) TakeScreenshot(req *computeruse.ScreenshotRequest) (*computeruse.ScreenshotResponse, error) {
	bounds := screenshot.GetDisplayBounds(0)
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return nil, err
	}

	// Convert to RGBA for drawing
	rgbaImg := image.NewRGBA(img.Bounds())
	draw.Draw(rgbaImg, rgbaImg.Bounds(), img, image.Point{}, draw.Src)

	// Draw cursor if requested
	mouseX, mouseY := 0, 0
	if req.ShowCursor {
		mouseX, mouseY = getMousePosition()
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

	return response, nil
}

// TakeRegionScreenshot takes a screenshot of a specific region
func (c *ComputerUse) TakeRegionScreenshot(req *computeruse.RegionScreenshotRequest) (*computeruse.ScreenshotResponse, error) {
	if err := validateScreenshotRegion(req.Width, req.Height); err != nil {
		return nil, err
	}

	rect := image.Rect(req.X, req.Y, req.X+req.Width, req.Y+req.Height)
	img, err := screenshot.CaptureRect(rect)
	if err != nil {
		return nil, fmt.Errorf("failed to capture screenshot region x=%d y=%d width=%d height=%d: %w", req.X, req.Y, req.Width, req.Height, err)
	}

	// Convert to RGBA for drawing
	rgbaImg := image.NewRGBA(img.Bounds())
	draw.Draw(rgbaImg, rgbaImg.Bounds(), img, image.Point{}, draw.Src)

	// Draw cursor if requested and it's within the region
	mouseX, mouseY := 0, 0
	if req.ShowCursor {
		absoluteMouseX, absoluteMouseY := getMousePosition()
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

	return response, nil
}

// TakeCompressedScreenshot takes a compressed screenshot of the entire screen
func (c *ComputerUse) TakeCompressedScreenshot(req *computeruse.CompressedScreenshotRequest) (*computeruse.ScreenshotResponse, error) {
	params := ImageCompressionParams{
		Format:  req.Format,
		Quality: req.Quality,
		Scale:   req.Scale,
	}

	bounds := screenshot.GetDisplayBounds(0)
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return nil, err
	}

	// Convert to RGBA for drawing
	rgbaImg := image.NewRGBA(img.Bounds())
	draw.Draw(rgbaImg, rgbaImg.Bounds(), img, image.Point{}, draw.Src)

	// Draw cursor if requested
	mouseX, mouseY := 0, 0
	if req.ShowCursor {
		mouseX, mouseY = getMousePosition()
		drawCursor(rgbaImg, mouseX, mouseY)
	}

	// Encode with compression
	imageData, err := encodeImageWithCompression(rgbaImg, params)
	if err != nil {
		return nil, err
	}

	base64Str := base64.StdEncoding.EncodeToString(imageData)

	response := &computeruse.ScreenshotResponse{
		Screenshot: base64Str,
		SizeBytes:  len(imageData),
	}

	if req.ShowCursor {
		response.CursorPosition = &computeruse.Position{
			X: int(float64(mouseX) * params.Scale),
			Y: int(float64(mouseY) * params.Scale),
		}
	}

	return response, nil
}

// TakeCompressedRegionScreenshot takes a compressed screenshot of a specific region
func (c *ComputerUse) TakeCompressedRegionScreenshot(req *computeruse.CompressedRegionScreenshotRequest) (*computeruse.ScreenshotResponse, error) {
	params := ImageCompressionParams{
		Format:  req.Format,
		Quality: req.Quality,
		Scale:   req.Scale,
	}

	if err := validateScreenshotRegion(req.Width, req.Height); err != nil {
		return nil, err
	}

	rect := image.Rect(req.X, req.Y, req.X+req.Width, req.Y+req.Height)
	img, err := screenshot.CaptureRect(rect)
	if err != nil {
		return nil, fmt.Errorf("failed to capture screenshot region x=%d y=%d width=%d height=%d: %w", req.X, req.Y, req.Width, req.Height, err)
	}

	// Convert to RGBA for drawing
	rgbaImg := image.NewRGBA(img.Bounds())
	draw.Draw(rgbaImg, rgbaImg.Bounds(), img, image.Point{}, draw.Src)

	// Draw cursor if requested and it's within the region
	mouseX, mouseY := 0, 0
	if req.ShowCursor {
		absoluteMouseX, absoluteMouseY := getMousePosition()
		mouseX = absoluteMouseX - req.X
		mouseY = absoluteMouseY - req.Y

		if mouseX >= 0 && mouseX < req.Width && mouseY >= 0 && mouseY < req.Height {
			drawCursor(rgbaImg, mouseX, mouseY)
		}
	}

	// Encode with compression
	imageData, err := encodeImageWithCompression(rgbaImg, params)
	if err != nil {
		return nil, err
	}

	base64Str := base64.StdEncoding.EncodeToString(imageData)

	response := &computeruse.ScreenshotResponse{
		Screenshot: base64Str,
		SizeBytes:  len(imageData),
	}

	if req.ShowCursor {
		response.CursorPosition = &computeruse.Position{
			X: req.X + int(float64(mouseX)*params.Scale),
			Y: req.Y + int(float64(mouseY)*params.Scale),
		}
	}

	return response, nil
}

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
