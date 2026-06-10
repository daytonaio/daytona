// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
)

// ImageCompressionParams holds parameters for image compression
type ImageCompressionParams struct {
	Format  string  // "png" or "jpeg"
	Quality int     // 1-100 for JPEG quality
	Scale   float64 // 0.1-1.0 for scaling down
}

func validateScreenshotRegion(width, height int) error {
	if width <= 0 || height <= 0 {
		return fmt.Errorf("invalid screenshot region: width and height must be greater than zero")
	}

	return nil
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
