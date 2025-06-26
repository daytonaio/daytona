// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

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

// ImageCompressionParams holds parameters for image compression
type ImageCompressionParams struct {
	Format  string  // "png", "jpeg", "webp"
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
	newWidth := int(float64(oldBounds.Dx()) * scale)
	newHeight := int(float64(oldBounds.Dy()) * scale)

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
		mouseX, mouseY = robotgo.Location()
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
		absoluteMouseX, absoluteMouseY := robotgo.Location()
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

func (u *ComputerUse) GetDisplayInfo() (*computeruse.DisplayInfoResponse, error) {
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
				ID:    id,
				Title: title,
				Position: computeruse.Position{
					X: 0, // Would need platform-specific implementation
					Y: 0, // Would need platform-specific implementation
				},
				Size: computeruse.Size{
					Width:  0, // Would need platform-specific implementation
					Height: 0, // Would need platform-specific implementation
				},
				IsActive: false, // Would need platform-specific implementation
			})
		}
	}

	return &computeruse.WindowsResponse{
		Windows: windows,
	}, nil
}
