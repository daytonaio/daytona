//go:build no_gui
// +build no_gui

// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"

	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
)

// Stub implementations for when GUI automation is not available

func (u *ComputerUse) TakeScreenshot(req *computeruse.ScreenshotRequest) (*computeruse.ScreenshotResponse, error) {
	// Create a mock screenshot (1x1 pixel)
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{128, 128, 128, 255}) // Gray pixel

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}

	base64Str := base64.StdEncoding.EncodeToString(buf.Bytes())

	return &computeruse.ScreenshotResponse{
		Screenshot: base64Str,
		Width:      1,
		Height:     1,
	}, nil
}

func (u *ComputerUse) TakeRegionScreenshot(req *computeruse.RegionScreenshotRequest) (*computeruse.ScreenshotResponse, error) {
	// Create a mock region screenshot
	img := image.NewRGBA(image.Rect(0, 0, req.Width, req.Height))
	for y := 0; y < req.Height; y++ {
		for x := 0; x < req.Width; x++ {
			img.Set(x, y, color.RGBA{128, 128, 128, 255}) // Gray pixels
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}

	base64Str := base64.StdEncoding.EncodeToString(buf.Bytes())

	return &computeruse.ScreenshotResponse{
		Screenshot: base64Str,
		Width:      req.Width,
		Height:     req.Height,
		Region:     req,
	}, nil
}

func (u *ComputerUse) TakeCompressedScreenshot(req *computeruse.CompressedScreenshotRequest) (*computeruse.ScreenshotResponse, error) {
	// Create a mock compressed screenshot
	img := image.NewRGBA(image.Rect(0, 0, 1920, 1080))
	for y := 0; y < 1080; y++ {
		for x := 0; x < 1920; x++ {
			img.Set(x, y, color.RGBA{128, 128, 128, 255}) // Gray pixels
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}

	base64Str := base64.StdEncoding.EncodeToString(buf.Bytes())

	return &computeruse.ScreenshotResponse{
		Screenshot: base64Str,
		Width:      1920,
		Height:     1080,
		Format:     req.Format,
		Quality:    req.Quality,
		Scale:      req.Scale,
		SizeBytes:  len(buf.Bytes()),
	}, nil
}

func (u *ComputerUse) TakeCompressedRegionScreenshot(req *computeruse.CompressedRegionScreenshotRequest) (*computeruse.ScreenshotResponse, error) {
	// Create a mock compressed region screenshot
	img := image.NewRGBA(image.Rect(0, 0, req.Width, req.Height))
	for y := 0; y < req.Height; y++ {
		for x := 0; x < req.Width; x++ {
			img.Set(x, y, color.RGBA{128, 128, 128, 255}) // Gray pixels
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}

	base64Str := base64.StdEncoding.EncodeToString(buf.Bytes())

	region := &computeruse.RegionScreenshotRequest{
		X:          req.X,
		Y:          req.Y,
		Width:      req.Width,
		Height:     req.Height,
		ShowCursor: req.ShowCursor,
	}

	return &computeruse.ScreenshotResponse{
		Screenshot: base64Str,
		Width:      req.Width,
		Height:     req.Height,
		Region:     region,
		Format:     req.Format,
		Quality:    req.Quality,
		Scale:      req.Scale,
		SizeBytes:  len(buf.Bytes()),
	}, nil
}

func (u *ComputerUse) GetMousePosition() (*computeruse.MousePositionResponse, error) {
	return &computeruse.MousePositionResponse{
		X: 0,
		Y: 0,
	}, nil
}

func (u *ComputerUse) MoveMouse(req *computeruse.MoveMouseRequest) (*computeruse.MousePositionResponse, error) {
	return &computeruse.MousePositionResponse{
		Success: true,
		X:       req.X,
		Y:       req.Y,
		ActualX: req.X,
		ActualY: req.Y,
	}, nil
}

func (u *ComputerUse) Click(req *computeruse.ClickRequest) (*computeruse.MousePositionResponse, error) {
	if req.Button == "" {
		req.Button = "left"
	}

	return &computeruse.MousePositionResponse{
		Success: true,
		Action:  "click",
		Button:  req.Button,
		Double:  req.Double,
		X:       req.X,
		Y:       req.Y,
		ActualX: req.X,
		ActualY: req.Y,
	}, nil
}

func (u *ComputerUse) Drag(req *computeruse.DragRequest) (*computeruse.MousePositionResponse, error) {
	if req.Button == "" {
		req.Button = "left"
	}

	return &computeruse.MousePositionResponse{
		Success: true,
		Action:  "drag",
		From: &computeruse.MousePositionResponse{
			X: req.StartX,
			Y: req.StartY,
		},
		To: &computeruse.MousePositionResponse{
			X: req.EndX,
			Y: req.EndY,
		},
		ActualX: req.EndX,
		ActualY: req.EndY,
	}, nil
}

func (u *ComputerUse) Scroll(req *computeruse.ScrollRequest) (*computeruse.Empty, error) {
	return new(computeruse.Empty), nil
}

func (u *ComputerUse) TypeText(req *computeruse.TypeTextRequest) (*computeruse.Empty, error) {
	return new(computeruse.Empty), nil
}

func (u *ComputerUse) PressKey(req *computeruse.PressKeyRequest) (*computeruse.Empty, error) {
	return new(computeruse.Empty), nil
}

func (u *ComputerUse) PressHotkey(req *computeruse.PressHotkeyRequest) (*computeruse.Empty, error) {
	return new(computeruse.Empty), nil
}

func (u *ComputerUse) GetDisplayInfo() (*computeruse.DisplayInfoResponse, error) {
	return &computeruse.DisplayInfoResponse{
		Displays: []computeruse.DisplayInfo{
			{
				ID:       0,
				X:        0,
				Y:        0,
				Width:    1920,
				Height:   1080,
				IsActive: true,
			},
		},
	}, nil
}

func (u *ComputerUse) GetWindows() (*computeruse.WindowsResponse, error) {
	return &computeruse.WindowsResponse{
		Windows: []computeruse.WindowInfo{
			{
				ID:       1,
				Title:    "Mock Window",
				X:        0,
				Y:        0,
				Width:    800,
				Height:   600,
				IsActive: true,
			},
		},
	}, nil
}
