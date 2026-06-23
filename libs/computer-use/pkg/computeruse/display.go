// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"

	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
)

// ImageCompressionParams holds parameters for image compression
type ImageCompressionParams struct {
	Format  string  // "png" or "jpeg"
	Quality int     // 1-100 for JPEG quality
	Scale   float64 // 0.1-1.0 for scaling down
}

// encodeImageWithCompression encodes an image with the specified compression parameters
func encodeImageWithCompression(img *image.RGBA, params ImageCompressionParams) ([]byte, error) {
	var buf bytes.Buffer

	switch params.Format {
	case "jpeg":
		// Scale the image if needed
		var scaledImg image.Image = img
		if params.Scale != 1.0 {
			scaledImg = scaleImage(img, params.Scale)
		}
		err := jpeg.Encode(&buf, scaledImg, &jpeg.Options{Quality: params.Quality})
		if err != nil {
			return nil, err
		}
	case "png":
		// Scale the image if needed
		var scaledImg image.Image = img
		if params.Scale != 1.0 {
			scaledImg = scaleImage(img, params.Scale)
		}
		err := png.Encode(&buf, scaledImg)
		if err != nil {
			return nil, err
		}
	default:
		// Default to PNG
		err := png.Encode(&buf, img)
		if err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// scaleImage scales an image by the given factor
func scaleImage(img *image.RGBA, scale float64) image.Image {
	if scale == 1.0 {
		return img
	}

	// Simple nearest neighbor scaling
	oldBounds := img.Bounds()
	if oldBounds.Dx() <= 0 || oldBounds.Dy() <= 0 {
		return img
	}

	newWidth := int(float64(oldBounds.Dx()) * scale)
	newHeight := int(float64(oldBounds.Dy()) * scale)
	if newWidth < 1 {
		newWidth = 1
	}
	if newHeight < 1 {
		newHeight = 1
	}

	scaledImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			oldX := int(float64(x) / scale)
			oldY := int(float64(y) / scale)
			if oldX < oldBounds.Dx() && oldY < oldBounds.Dy() {
				scaledImg.Set(x, y, img.At(oldX, oldY))
			}
		}
	}

	return scaledImg
}

func (u *ComputerUse) TakeCompressedScreenshot(req *computeruse.CompressedScreenshotRequest) (*computeruse.ScreenshotResponse, error) {
	params := ImageCompressionParams{
		Format:  req.Format,
		Quality: req.Quality,
		Scale:   req.Scale,
	}

	img, mouseX, mouseY, err := captureWithCursor(req.ShowCursor, 0, 0, capturePrimaryDisplay)
	if err != nil {
		return nil, err
	}

	return encodeCompressedScreenshot(img, params, req.ShowCursor, mouseX, mouseY, 0, 0)
}

// TakeCompressedRegionScreenshot takes a region screenshot with compression options
func (u *ComputerUse) TakeCompressedRegionScreenshot(req *computeruse.CompressedRegionScreenshotRequest) (*computeruse.ScreenshotResponse, error) {
	params := ImageCompressionParams{
		Format:  req.Format,
		Quality: req.Quality,
		Scale:   req.Scale,
	}

	if err := validateScreenshotRegion(req.Width, req.Height); err != nil {
		return nil, err
	}

	rect := image.Rect(req.X, req.Y, req.X+req.Width, req.Y+req.Height)
	img, mouseX, mouseY, err := captureWithCursor(req.ShowCursor, req.X, req.Y, func() (*image.RGBA, error) {
		return captureRect(rect)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to capture screenshot region x=%d y=%d width=%d height=%d: %w", req.X, req.Y, req.Width, req.Height, err)
	}

	return encodeCompressedScreenshot(img, params, req.ShowCursor, mouseX, mouseY, req.X, req.Y)
}

func encodeCompressedScreenshot(img image.Image, params ImageCompressionParams, showCursor bool, mouseX, mouseY, offsetX, offsetY int) (*computeruse.ScreenshotResponse, error) {
	// Convert to RGBA for drawing
	rgbaImg := image.NewRGBA(img.Bounds())
	draw.Draw(rgbaImg, rgbaImg.Bounds(), img, image.Point{}, draw.Src)

	if showCursor && mouseX >= 0 && mouseX < img.Bounds().Dx() && mouseY >= 0 && mouseY < img.Bounds().Dy() {
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

	if showCursor {
		response.CursorPosition = &computeruse.Position{
			X: offsetX + int(float64(mouseX)*params.Scale),
			Y: offsetY + int(float64(mouseY)*params.Scale),
		}
	}

	return response, nil
}

func (u *ComputerUse) GetDisplayInfo() (*computeruse.DisplayInfoResponse, error) {
	var displays []computeruse.DisplayInfo
	err := withX11Client(func() error {
		var err error
		displays, err = getDisplayInfos()
		return err
	})
	if err != nil {
		return nil, err
	}

	return &computeruse.DisplayInfoResponse{
		Displays: displays,
	}, nil
}

func (u *ComputerUse) GetWindows() (*computeruse.WindowsResponse, error) {
	windows := make([]computeruse.WindowInfo, 0)
	err := withX11Client(func() error {
		var err error
		windows, err = getWindows()
		return err
	})
	if err != nil {
		return nil, err
	}

	return &computeruse.WindowsResponse{
		Windows: windows,
	}, nil
}
