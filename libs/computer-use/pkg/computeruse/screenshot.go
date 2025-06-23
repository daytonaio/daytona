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
	"net/http"

	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/gin-gonic/gin"
	"github.com/go-vgo/robotgo"
	"github.com/kbinani/screenshot"
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

func (u *ComputerUse) TakeScreenshot(req *computeruse.ComputerUseRequest) (*computeruse.Empty, error) {
	// Check if we should show cursor
	showCursor := req.RequestContext.Query("show_cursor") == "true"

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

	// Convert to base64
	var buf bytes.Buffer
	if err := png.Encode(&buf, rgbaImg); err != nil {
		req.RequestContext.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to encode image",
		})
		return new(computeruse.Empty), nil
	}

	base64Str := base64.StdEncoding.EncodeToString(buf.Bytes())

	response := gin.H{
		"screenshot": base64Str,
		"width":      bounds.Dx(),
		"height":     bounds.Dy(),
	}

	if showCursor {
		response["cursor_position"] = gin.H{
			"x": mouseX,
			"y": mouseY,
		}
	}

	req.RequestContext.JSON(http.StatusOK, response)

	return new(computeruse.Empty), nil
}

func (u *ComputerUse) TakeRegionScreenshot(req *computeruse.ComputerUseRequest) (*computeruse.Empty, error) {
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
		// Convert to relative coordinates within the region
		mouseX = absoluteMouseX - region.X
		mouseY = absoluteMouseY - region.Y

		// Only draw if cursor is within the region
		if mouseX >= 0 && mouseX < region.Width && mouseY >= 0 && mouseY < region.Height {
			drawCursor(rgbaImg, mouseX, mouseY)
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, rgbaImg); err != nil {
		req.RequestContext.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to encode image",
		})
		return new(computeruse.Empty), nil
	}

	base64Str := base64.StdEncoding.EncodeToString(buf.Bytes())

	response := gin.H{
		"screenshot": base64Str,
		"region":     region,
	}

	if showCursor {
		response["cursor_position"] = gin.H{
			"x": mouseX + region.X,
			"y": mouseY + region.Y,
		}
	}

	req.RequestContext.JSON(http.StatusOK, response)

	return new(computeruse.Empty), nil
}
