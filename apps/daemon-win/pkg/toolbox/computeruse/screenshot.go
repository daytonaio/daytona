// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"

	"github.com/kbinani/screenshot"
)

// ImageCompressionParams holds parameters for image compression
type ImageCompressionParams struct {
	Format  string  // "png", "jpeg"
	Quality int     // 1-100 for JPEG quality
	Scale   float64 // 0.1-1.0 for scaling down
}

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

// TakeScreenshot takes a screenshot of the entire screen
func (c *ComputerUse) TakeScreenshot(req *ScreenshotRequest) (*ScreenshotResponse, error) {
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

	response := &ScreenshotResponse{
		Screenshot: base64Str,
	}

	if req.ShowCursor {
		response.CursorPosition = &Position{
			X: mouseX,
			Y: mouseY,
		}
	}

	return response, nil
}

// TakeRegionScreenshot takes a screenshot of a specific region
func (c *ComputerUse) TakeRegionScreenshot(req *RegionScreenshotRequest) (*ScreenshotResponse, error) {
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

	response := &ScreenshotResponse{
		Screenshot: base64Str,
	}

	if req.ShowCursor {
		response.CursorPosition = &Position{
			X: mouseX + req.X,
			Y: mouseY + req.Y,
		}
	}

	return response, nil
}

// TakeCompressedScreenshot takes a compressed screenshot of the entire screen
func (c *ComputerUse) TakeCompressedScreenshot(req *CompressedScreenshotRequest) (*ScreenshotResponse, error) {
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

	response := &ScreenshotResponse{
		Screenshot: base64Str,
		SizeBytes:  len(imageData),
	}

	if req.ShowCursor {
		response.CursorPosition = &Position{
			X: int(float64(mouseX) * params.Scale),
			Y: int(float64(mouseY) * params.Scale),
		}
	}

	return response, nil
}

// TakeCompressedRegionScreenshot takes a compressed screenshot of a specific region
func (c *ComputerUse) TakeCompressedRegionScreenshot(req *CompressedRegionScreenshotRequest) (*ScreenshotResponse, error) {
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

	response := &ScreenshotResponse{
		Screenshot: base64Str,
		SizeBytes:  len(imageData),
	}

	if req.ShowCursor {
		response.CursorPosition = &Position{
			X: req.X + int(float64(mouseX)*params.Scale),
			Y: req.Y + int(float64(mouseY)*params.Scale),
		}
	}

	return response, nil
}
