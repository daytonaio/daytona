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
	Click(*ClickRequest) (*MousePositionResponse, error)
	Drag(*DragRequest) (*MousePositionResponse, error)
	Scroll(*ScrollRequest) (*Empty, error)

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

// Screenshot parameter structs
type ScreenshotRequest struct {
	ShowCursor bool `json:"show_cursor"`
}

type RegionScreenshotRequest struct {
	X          int  `json:"x"`
	Y          int  `json:"y"`
	Width      int  `json:"width"`
	Height     int  `json:"height"`
	ShowCursor bool `json:"show_cursor"`
}

type CompressedScreenshotRequest struct {
	ShowCursor bool    `json:"show_cursor"`
	Format     string  `json:"format"`  // "png" or "jpeg"
	Quality    int     `json:"quality"` // 1-100 for JPEG quality
	Scale      float64 `json:"scale"`   // 0.1-1.0 for scaling down
}

type CompressedRegionScreenshotRequest struct {
	X          int     `json:"x"`
	Y          int     `json:"y"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	ShowCursor bool    `json:"show_cursor"`
	Format     string  `json:"format"`  // "png" or "jpeg"
	Quality    int     `json:"quality"` // 1-100 for JPEG quality
	Scale      float64 `json:"scale"`   // 0.1-1.0 for scaling down
}

// Mouse parameter structs
type MoveMouseRequest struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type ClickRequest struct {
	X      int    `json:"x"`
	Y      int    `json:"y"`
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
	X         int    `json:"x"`
	Y         int    `json:"y"`
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

// Response structs
type ScreenshotResponse struct {
	Screenshot     string                   `json:"screenshot"`
	Width          int                      `json:"width"`
	Height         int                      `json:"height"`
	CursorPosition *MousePositionResponse   `json:"cursor_position,omitempty"`
	Region         *RegionScreenshotRequest `json:"region,omitempty"`
	Format         string                   `json:"format,omitempty"`
	Quality        int                      `json:"quality,omitempty"`
	Scale          float64                  `json:"scale,omitempty"`
	SizeBytes      int                      `json:"size_bytes,omitempty"`
}

type MousePositionResponse struct {
	X       int                    `json:"x"`
	Y       int                    `json:"y"`
	ActualX int                    `json:"actual_x,omitempty"`
	ActualY int                    `json:"actual_y,omitempty"`
	Success bool                   `json:"success,omitempty"`
	Action  string                 `json:"action,omitempty"`
	Button  string                 `json:"button,omitempty"`
	Double  bool                   `json:"double,omitempty"`
	From    *MousePositionResponse `json:"from,omitempty"`
	To      *MousePositionResponse `json:"to,omitempty"`
}

type DisplayInfoResponse struct {
	Displays []DisplayInfo `json:"displays"`
}

type DisplayInfo struct {
	ID       int  `json:"id"`
	X        int  `json:"x"`
	Y        int  `json:"y"`
	Width    int  `json:"width"`
	Height   int  `json:"height"`
	IsActive bool `json:"is_active"`
}

type WindowsResponse struct {
	Windows []WindowInfo `json:"windows"`
}

type WindowInfo struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	IsActive bool   `json:"is_active"`
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
			ShowCursor: c.Query("show_cursor") == "true",
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
		req.ShowCursor = c.Query("show_cursor") == "true"

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
			ShowCursor: c.Query("show_cursor") == "true",
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
		req.ShowCursor = c.Query("show_cursor") == "true"
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

func WrapClickHandler(fn func(*ClickRequest) (*MousePositionResponse, error)) gin.HandlerFunc {
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

func WrapDragHandler(fn func(*DragRequest) (*MousePositionResponse, error)) gin.HandlerFunc {
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

func WrapScrollHandler(fn func(*ScrollRequest) (*Empty, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ScrollRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scroll parameters"})
			return
		}

		_, err := fn(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}

func WrapTypeTextHandler(fn func(*TypeTextRequest) (*Empty, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req TypeTextRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		_, err := fn(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "typed": req.Text})
	}
}

func WrapPressKeyHandler(fn func(*PressKeyRequest) (*Empty, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req PressKeyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid key"})
			return
		}

		_, err := fn(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "key": req.Key, "modifiers": req.Modifiers})
	}
}

func WrapPressHotkeyHandler(fn func(*PressHotkeyRequest) (*Empty, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req PressHotkeyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hotkey"})
			return
		}

		_, err := fn(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "hotkey": req.Keys})
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
