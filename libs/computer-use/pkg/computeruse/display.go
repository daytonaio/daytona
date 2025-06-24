// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build !no_gui
// +build !no_gui

package computeruse

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"

	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/go-vgo/robotgo"
	"github.com/kbinani/screenshot"
)

// ImageCompressionParams holds compression settings
type ImageCompressionParams struct {
	Format  string  `form:"format" json:"format"`   // "png" or "jpeg"
	Quality int     `form:"quality" json:"quality"` // 1-100 for JPEG quality
	Scale   float64 `form:"scale" json:"scale"`     // 0.1-1.0 for scaling down
}

// encodeImageWithCompression encodes an image with the specified compression settings
func encodeImageWithCompression(img image.Image, params ImageCompressionParams) ([]byte, error) {
	var buf bytes.Buffer

	// Scale image if needed
	if params.Scale < 1.0 {
		bounds := img.Bounds()
		newWidth := int(float64(bounds.Dx()) * params.Scale)
		newHeight := int(float64(bounds.Dy()) * params.Scale)

		// Create scaled image
		scaledImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

		// Simple nearest neighbor scaling
		for y := 0; y < newHeight; y++ {
			for x := 0; x < newWidth; x++ {
				srcX := int(float64(x) / params.Scale)
				srcY := int(float64(y) / params.Scale)
				scaledImg.Set(x, y, img.At(srcX, srcY))
			}
		}
		img = scaledImg
	}

	// Encode based on format
	switch params.Format {
	case "jpeg":
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: params.Quality})
		return buf.Bytes(), err
	default: // png
		err := png.Encode(&buf, img)
		return buf.Bytes(), err
	}
}

// TakeCompressedScreenshot takes a screenshot with compression options
func (u *ComputerUse) TakeCompressedScreenshot(req *computeruse.CompressedScreenshotRequest) (*computeruse.ScreenshotResponse, error) {
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
		mouseX, mouseY = robotgo.GetMousePos()
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
		Width:      int(float64(bounds.Dx()) * params.Scale),
		Height:     int(float64(bounds.Dy()) * params.Scale),
		Format:     params.Format,
		Quality:    params.Quality,
		Scale:      params.Scale,
		SizeBytes:  len(imageData),
	}

	if req.ShowCursor {
		response.CursorPosition = &computeruse.MousePositionResponse{
			X: int(float64(mouseX) * params.Scale),
			Y: int(float64(mouseY) * params.Scale),
		}
	}

	return response, nil
}

// TakeCompressedRegionScreenshot takes a region screenshot with compression options
func (u *ComputerUse) TakeCompressedRegionScreenshot(req *computeruse.CompressedRegionScreenshotRequest) (*computeruse.ScreenshotResponse, error) {
	params := ImageCompressionParams{
		Format:  req.Format,
		Quality: req.Quality,
		Scale:   req.Scale,
	}

	rect := image.Rect(req.X, req.Y, req.X+req.Width, req.Y+req.Height)
	img, err := screenshot.CaptureRect(rect)
	if err != nil {
		return nil, err
	}

	// Convert to RGBA for drawing
	rgbaImg := image.NewRGBA(img.Bounds())
	draw.Draw(rgbaImg, rgbaImg.Bounds(), img, image.Point{}, draw.Src)

	// Draw cursor if requested and it's within the region
	mouseX, mouseY := 0, 0
	if req.ShowCursor {
		absoluteMouseX, absoluteMouseY := robotgo.GetMousePos()
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

	region := &computeruse.RegionScreenshotRequest{
		X:          req.X,
		Y:          req.Y,
		Width:      int(float64(req.Width) * params.Scale),
		Height:     int(float64(req.Height) * params.Scale),
		ShowCursor: req.ShowCursor,
	}

	response := &computeruse.ScreenshotResponse{
		Screenshot: base64Str,
		Width:      int(float64(req.Width) * params.Scale),
		Height:     int(float64(req.Height) * params.Scale),
		Region:     region,
		Format:     params.Format,
		Quality:    params.Quality,
		Scale:      params.Scale,
		SizeBytes:  len(imageData),
	}

	if req.ShowCursor {
		response.CursorPosition = &computeruse.MousePositionResponse{
			X: req.X + int(float64(mouseX)*params.Scale),
			Y: req.Y + int(float64(mouseY)*params.Scale),
		}
	}

	return response, nil
}

func (u *ComputerUse) GetDisplayInfo() (*computeruse.DisplayInfoResponse, error) {
	n := screenshot.NumActiveDisplays()
	displays := make([]computeruse.DisplayInfo, n)

	for i := 0; i < n; i++ {
		bounds := screenshot.GetDisplayBounds(i)
		displays[i] = computeruse.DisplayInfo{
			ID:       i,
			X:        bounds.Min.X,
			Y:        bounds.Min.Y,
			Width:    bounds.Dx(),
			Height:   bounds.Dy(),
			IsActive: true, // Assuming all detected displays are active
		}
	}

	return &computeruse.DisplayInfoResponse{
		Displays: displays,
	}, nil
}

func (u *ComputerUse) GetWindows() (*computeruse.WindowsResponse, error) {
	// This is a simplified version - robotgo's window functions
	// might need additional setup depending on the platform

	titles, err := robotgo.FindIds("")
	if err != nil {
		return nil, err
	}

	windows := make([]computeruse.WindowInfo, 0)
	for _, id := range titles {
		title := robotgo.GetTitle(id)
		if title != "" {
			// Get window position and size (this might need platform-specific implementation)
			// For now, we'll use placeholder values
			windows = append(windows, computeruse.WindowInfo{
				ID:       id,
				Title:    title,
				X:        0,     // Would need platform-specific implementation
				Y:        0,     // Would need platform-specific implementation
				Width:    0,     // Would need platform-specific implementation
				Height:   0,     // Would need platform-specific implementation
				IsActive: false, // Would need platform-specific implementation
			})
		}
	}

	return &computeruse.WindowsResponse{
		Windows: windows,
	}, nil
}
