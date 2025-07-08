// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"net/http"
	"net/rpc"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-plugin"
)

// PluginInterface defines the interface that the computeruse plugin must implement
type IComputerUse interface {
	// Process management
	Initialize() (*Empty, error)
	Start() (*Empty, error)
	Stop() (*Empty, error)
	GetProcessStatus() (map[string]ProcessStatus, error)
	IsProcessRunning(req *ProcessRequest) (bool, error)
	RestartProcess(req *ProcessRequest) (*Empty, error)
	GetProcessLogs(req *ProcessRequest) (string, error)
	GetProcessErrors(req *ProcessRequest) (string, error)

	// Screenshot methods
	TakeScreenshot(*ScreenshotRequest) (*ScreenshotResponse, error)
	TakeRegionScreenshot(*RegionScreenshotRequest) (*ScreenshotResponse, error)
	TakeCompressedScreenshot(*CompressedScreenshotRequest) (*ScreenshotResponse, error)
	TakeCompressedRegionScreenshot(*CompressedRegionScreenshotRequest) (*ScreenshotResponse, error)

	// Mouse control methods
	GetMousePosition() (*MousePositionResponse, error)
	MoveMouse(*MoveMouseRequest) (*MousePositionResponse, error)
	Click(*ClickRequest) (*MouseClickResponse, error)
	Drag(*DragRequest) (*MouseDragResponse, error)
	Scroll(*ScrollRequest) (*ScrollResponse, error)

	// Keyboard control methods
	TypeText(*TypeTextRequest) (*Empty, error)
	PressKey(*PressKeyRequest) (*Empty, error)
	PressHotkey(*PressHotkeyRequest) (*Empty, error)

	// Display info methods
	GetDisplayInfo() (*DisplayInfoResponse, error)
	GetWindows() (*WindowsResponse, error)

	// Status method
	GetStatus() (*StatusResponse, error)
}

type ComputerUsePlugin struct {
	Impl IComputerUse
}

// Common structs for better composition
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Size struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Screenshot parameter structs
type ScreenshotRequest struct {
	ShowCursor bool `json:"showCursor"`
}

type RegionScreenshotRequest struct {
	Position
	Size
	ShowCursor bool `json:"showCursor"`
}

type CompressedScreenshotRequest struct {
	ShowCursor bool    `json:"showCursor"`
	Format     string  `json:"format"`  // "png" or "jpeg"
	Quality    int     `json:"quality"` // 1-100 for JPEG quality
	Scale      float64 `json:"scale"`   // 0.1-1.0 for scaling down
}

type CompressedRegionScreenshotRequest struct {
	Position
	Size
	ShowCursor bool    `json:"showCursor"`
	Format     string  `json:"format"`  // "png" or "jpeg"
	Quality    int     `json:"quality"` // 1-100 for JPEG quality
	Scale      float64 `json:"scale"`   // 0.1-1.0 for scaling down
}

// Mouse parameter structs
type MoveMouseRequest struct {
	Position
}

type ClickRequest struct {
	Position
	Button string `json:"button"` // left, right, middle
	Double bool   `json:"double"`
}

type DragRequest struct {
	StartX int    `json:"startX"`
	StartY int    `json:"startY"`
	EndX   int    `json:"endX"`
	EndY   int    `json:"endY"`
	Button string `json:"button"`
}

type ScrollRequest struct {
	Position
	Direction string `json:"direction"` // up, down
	Amount    int    `json:"amount"`
}

// Keyboard parameter structs
type TypeTextRequest struct {
	Text  string `json:"text"`
	Delay int    `json:"delay"` // milliseconds between keystrokes
}

type PressKeyRequest struct {
	Key       string   `json:"key"`
	Modifiers []string `json:"modifiers"` // ctrl, alt, shift, cmd
}

type PressHotkeyRequest struct {
	Keys string `json:"keys"` // e.g., "ctrl+c", "cmd+v"
}

// Response structs for keyboard operations
type ScrollResponse struct {
	Success bool `json:"success"`
}

// Response structs
type ScreenshotResponse struct {
	Screenshot     string    `json:"screenshot"`
	CursorPosition *Position `json:"cursorPosition,omitempty"`
	SizeBytes      int       `json:"sizeBytes,omitempty"`
}

// Mouse response structs - separated by operation type
type MousePositionResponse struct {
	Position
}

type MouseClickResponse struct {
	Position
}

type MouseDragResponse struct {
	Position // Final position
}

type DisplayInfoResponse struct {
	Displays []DisplayInfo `json:"displays"`
}

type DisplayInfo struct {
	ID int `json:"id"`
	Position
	Size
	IsActive bool `json:"isActive"`
}

type WindowsResponse struct {
	Windows []WindowInfo `json:"windows"`
}

type WindowInfo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Position
	Size
	IsActive bool `json:"isActive"`
}

type StatusResponse struct {
	Status string `json:"status"`
}

type ProcessStatus struct {
	Running     bool
	Priority    int
	AutoRestart bool
	Pid         *int
}

type ProcessRequest struct {
	ProcessName string
}

type Empty struct{}

func (p *ComputerUsePlugin) Server(*plugin.MuxBroker) (any, error) {
	return &ComputerUseRPCServer{Impl: p.Impl}, nil
}

func (p *ComputerUsePlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (any, error) {
	return &ComputerUseRPCClient{client: c}, nil
}

// Helper function to create handlers that convert gin context to specific request structs
func WrapScreenshotHandler(fn func(*ScreenshotRequest) (*ScreenshotResponse, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := &ScreenshotRequest{
			ShowCursor: c.Query("showCursor") == "true",
		}
		response, err := fn(req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	}
}

func WrapRegionScreenshotHandler(fn func(*RegionScreenshotRequest) (*ScreenshotResponse, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegionScreenshotRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parameters"})
			return
		}
		req.ShowCursor = c.Query("showCursor") == "true"

		response, err := fn(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	}
}

func WrapCompressedScreenshotHandler(fn func(*CompressedScreenshotRequest) (*ScreenshotResponse, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := &CompressedScreenshotRequest{
			ShowCursor: c.Query("showCursor") == "true",
			Format:     c.Query("format"),
			Quality:    85,
			Scale:      1.0,
		}

		// Parse quality
		if qualityStr := c.Query("quality"); qualityStr != "" {
			if quality, err := strconv.Atoi(qualityStr); err == nil && quality >= 1 && quality <= 100 {
				req.Quality = quality
			}
		}

		// Parse scale
		if scaleStr := c.Query("scale"); scaleStr != "" {
			if scale, err := strconv.ParseFloat(scaleStr, 64); err == nil && scale >= 0.1 && scale <= 1.0 {
				req.Scale = scale
			}
		}

		response, err := fn(req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	}
}

func WrapCompressedRegionScreenshotHandler(fn func(*CompressedRegionScreenshotRequest) (*ScreenshotResponse, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CompressedRegionScreenshotRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parameters"})
			return
		}
		req.ShowCursor = c.Query("showCursor") == "true"
		req.Format = c.Query("format")
		req.Quality = 85
		req.Scale = 1.0

		// Parse quality
		if qualityStr := c.Query("quality"); qualityStr != "" {
			if quality, err := strconv.Atoi(qualityStr); err == nil && quality >= 1 && quality <= 100 {
				req.Quality = quality
			}
		}

		// Parse scale
		if scaleStr := c.Query("scale"); scaleStr != "" {
			if scale, err := strconv.ParseFloat(scaleStr, 64); err == nil && scale >= 0.1 && scale <= 1.0 {
				req.Scale = scale
			}
		}

		response, err := fn(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	}
}

func WrapMousePositionHandler(fn func() (*MousePositionResponse, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		response, err := fn()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	}
}

func WrapMoveMouseHandler(fn func(*MoveMouseRequest) (*MousePositionResponse, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req MoveMouseRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coordinates"})
			return
		}

		response, err := fn(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	}
}

func WrapClickHandler(fn func(*ClickRequest) (*MouseClickResponse, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ClickRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid click parameters"})
			return
		}

		response, err := fn(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	}
}

func WrapDragHandler(fn func(*DragRequest) (*MouseDragResponse, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req DragRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid drag parameters"})
			return
		}

		response, err := fn(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	}
}

func WrapScrollHandler(fn func(*ScrollRequest) (*ScrollResponse, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ScrollRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scroll parameters"})
			return
		}

		response, err := fn(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	}
}

func WrapTypeTextHandler(fn func(*TypeTextRequest) (*Empty, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req TypeTextRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		response, err := fn(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	}
}

func WrapPressKeyHandler(fn func(*PressKeyRequest) (*Empty, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req PressKeyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid key"})
			return
		}

		response, err := fn(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	}
}

func WrapPressHotkeyHandler(fn func(*PressHotkeyRequest) (*Empty, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req PressHotkeyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hotkey"})
			return
		}

		response, err := fn(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	}
}

func WrapDisplayInfoHandler(fn func() (*DisplayInfoResponse, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		response, err := fn()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	}
}

func WrapWindowsHandler(fn func() (*WindowsResponse, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		response, err := fn()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	}
}

func WrapStatusHandler(fn func() (*StatusResponse, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		response, err := fn()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	}
}
