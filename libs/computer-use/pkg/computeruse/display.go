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
	"net/http"
	"strconv"

	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/gin-gonic/gin"
	"github.com/go-vgo/robotgo"
	"github.com/kbinani/screenshot"
)

// ImageCompressionParams holds compression settings
type ImageCompressionParams struct {
	Format  string  `form:"format" json:"format"`   // "png" or "jpeg"
	Quality int     `form:"quality" json:"quality"` // 1-100 for JPEG quality
	Scale   float64 `form:"scale" json:"scale"`     // 0.1-1.0 for scaling down
}

// getCompressionParams extracts and validates compression parameters from query
func getCompressionParams(c *gin.Context) ImageCompressionParams {
	params := ImageCompressionParams{
		Format:  "png",
		Quality: 85,
		Scale:   1.0,
	}

	// Parse format
	if format := c.Query("format"); format == "jpeg" || format == "jpg" {
		params.Format = "jpeg"
	}

	// Parse quality (for JPEG)
	if qualityStr := c.Query("quality"); qualityStr != "" {
		if quality, err := strconv.Atoi(qualityStr); err == nil {
			if quality >= 1 && quality <= 100 {
				params.Quality = quality
			}
		}
	}

	// Parse scale
	if scaleStr := c.Query("scale"); scaleStr != "" {
		if scale, err := strconv.ParseFloat(scaleStr, 64); err == nil {
			if scale >= 0.1 && scale <= 1.0 {
				params.Scale = scale
			}
		}
	}

	return params
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
func (u *ComputerUse) TakeCompressedScreenshot(req *computeruse.ComputerUseRequest) (*computeruse.Empty, error) {
	showCursor := req.RequestContext.Query("show_cursor") == "true"
	params := getCompressionParams(req.RequestContext)

	bounds := screenshot.GetDisplayBounds(0)
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		req.RequestContext.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to capture screenshot",
			"details": err.Error(),
		})
		return new(computeruse.Empty), nil
	}

	// Convert to RGBA for drawing
	rgbaImg := image.NewRGBA(img.Bounds())
	draw.Draw(rgbaImg, rgbaImg.Bounds(), img, image.Point{}, draw.Src)

	// Draw cursor if requested
	mouseX, mouseY := 0, 0
	if showCursor {
		mouseX, mouseY = robotgo.GetMousePos()
		drawCursor(rgbaImg, mouseX, mouseY)
	}

	// Encode with compression
	imageData, err := encodeImageWithCompression(rgbaImg, params)
	if err != nil {
		req.RequestContext.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to encode image",
		})
		return new(computeruse.Empty), nil
	}

	base64Str := base64.StdEncoding.EncodeToString(imageData)

	response := gin.H{
		"screenshot": base64Str,
		"width":      int(float64(bounds.Dx()) * params.Scale),
		"height":     int(float64(bounds.Dy()) * params.Scale),
		"format":     params.Format,
		"quality":    params.Quality,
		"scale":      params.Scale,
		"size_bytes": len(imageData),
	}

	if showCursor {
		response["cursor_position"] = gin.H{
			"x": int(float64(mouseX) * params.Scale),
			"y": int(float64(mouseY) * params.Scale),
		}
	}

	req.RequestContext.JSON(http.StatusOK, response)
	return new(computeruse.Empty), nil
}

// TakeCompressedRegionScreenshot takes a region screenshot with compression options
func (u *ComputerUse) TakeCompressedRegionScreenshot(req *computeruse.ComputerUseRequest) (*computeruse.Empty, error) {
	var region struct {
		X      int `form:"x"`
		Y      int `form:"y"`
		Width  int `form:"width"`
		Height int `form:"height"`
	}

	if err := req.RequestContext.ShouldBindQuery(&region); err != nil {
		req.RequestContext.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid parameters",
		})
		return new(computeruse.Empty), nil
	}

	showCursor := req.RequestContext.Query("show_cursor") == "true"
	params := getCompressionParams(req.RequestContext)

	rect := image.Rect(region.X, region.Y, region.X+region.Width, region.Y+region.Height)
	img, err := screenshot.CaptureRect(rect)
	if err != nil {
		req.RequestContext.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to capture region",
		})
		return new(computeruse.Empty), nil
	}

	// Convert to RGBA for drawing
	rgbaImg := image.NewRGBA(img.Bounds())
	draw.Draw(rgbaImg, rgbaImg.Bounds(), img, image.Point{}, draw.Src)

	// Draw cursor if requested and it's within the region
	mouseX, mouseY := 0, 0
	if showCursor {
		absoluteMouseX, absoluteMouseY := robotgo.GetMousePos()
		mouseX = absoluteMouseX - region.X
		mouseY = absoluteMouseY - region.Y

		if mouseX >= 0 && mouseX < region.Width && mouseY >= 0 && mouseY < region.Height {
			drawCursor(rgbaImg, mouseX, mouseY)
		}
	}

	// Encode with compression
	imageData, err := encodeImageWithCompression(rgbaImg, params)
	if err != nil {
		req.RequestContext.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to encode image",
		})
		return new(computeruse.Empty), nil
	}

	base64Str := base64.StdEncoding.EncodeToString(imageData)

	response := gin.H{
		"screenshot": base64Str,
		"region": gin.H{
			"x":      region.X,
			"y":      region.Y,
			"width":  int(float64(region.Width) * params.Scale),
			"height": int(float64(region.Height) * params.Scale),
		},
		"format":     params.Format,
		"quality":    params.Quality,
		"scale":      params.Scale,
		"size_bytes": len(imageData),
	}

	if showCursor {
		response["cursor_position"] = gin.H{
			"x": region.X + int(float64(mouseX)*params.Scale),
			"y": region.Y + int(float64(mouseY)*params.Scale),
		}
	}

	req.RequestContext.JSON(http.StatusOK, response)
	return new(computeruse.Empty), nil
}

func (u *ComputerUse) GetDisplayInfo(req *computeruse.ComputerUseRequest) (*computeruse.Empty, error) {
	n := screenshot.NumActiveDisplays()
	displays := make([]gin.H, n)

	for i := 0; i < n; i++ {
		bounds := screenshot.GetDisplayBounds(i)
		displays[i] = gin.H{
			"id":     i,
			"x":      bounds.Min.X,
			"y":      bounds.Min.Y,
			"width":  bounds.Dx(),
			"height": bounds.Dy(),
		}
	}

	sx, sy := robotgo.GetScreenSize()

	req.RequestContext.JSON(http.StatusOK, gin.H{
		"primary_display": gin.H{
			"width":  sx,
			"height": sy,
		},
		"displays":       displays,
		"total_displays": n,
	})
	return new(computeruse.Empty), nil
}

func (u *ComputerUse) GetWindows(req *computeruse.ComputerUseRequest) (*computeruse.Empty, error) {
	// This is a simplified version - robotgo's window functions
	// might need additional setup depending on the platform

	titles, err := robotgo.FindIds("")
	if err != nil {
		req.RequestContext.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get windows",
		})
		return new(computeruse.Empty), nil
	}

	windows := make([]gin.H, 0)
	for _, id := range titles {
		title := robotgo.GetTitle(id)
		if title != "" {
			windows = append(windows, gin.H{
				"id":    id,
				"title": title,
			})
		}
	}

	req.RequestContext.JSON(http.StatusOK, gin.H{
		"windows": windows,
		"count":   len(windows),
	})
	return new(computeruse.Empty), nil
}
