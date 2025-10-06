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
	MoveMouse(*MouseMoveRequest) (*MousePositionResponse, error)
	Click(*MouseClickRequest) (*MouseClickResponse, error)
	Drag(*MouseDragRequest) (*MouseDragResponse, error)
	Scroll(*MouseScrollRequest) (*ScrollResponse, error)

	// Keyboard control methods
	TypeText(*KeyboardTypeRequest) (*Empty, error)
	PressKey(*KeyboardPressRequest) (*Empty, error)
	PressHotkey(*KeyboardHotkeyRequest) (*Empty, error)

	// Display info methods
	GetDisplayInfo() (*DisplayInfoResponse, error)
	GetWindows() (*WindowsResponse, error)

	// Status method
	GetStatus() (*ComputerUseStatusResponse, error)
}

type ComputerUsePlugin struct {
	Impl IComputerUse
}

// Common structs for better composition
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
} //	@name	Position

type Size struct {
	Width  int `json:"width"`
	Height int `json:"height"`
} //	@name	Size

// Screenshot parameter structs
type ScreenshotRequest struct {
	ShowCursor bool `json:"showCursor"`
} //	@name	ScreenshotRequest

type RegionScreenshotRequest struct {
	Position
	Size
	ShowCursor bool `json:"showCursor"`
} //	@name	RegionScreenshotRequest

type CompressedScreenshotRequest struct {
	ShowCursor bool    `json:"showCursor"`
	Format     string  `json:"format"`  // "png" or "jpeg"
	Quality    int     `json:"quality"` // 1-100 for JPEG quality
	Scale      float64 `json:"scale"`   // 0.1-1.0 for scaling down
} //	@name	CompressedScreenshotRequest

type CompressedRegionScreenshotRequest struct {
	Position
	Size
	ShowCursor bool    `json:"showCursor"`
	Format     string  `json:"format"`  // "png" or "jpeg"
	Quality    int     `json:"quality"` // 1-100 for JPEG quality
	Scale      float64 `json:"scale"`   // 0.1-1.0 for scaling down
} //	@name	CompressedRegionScreenshotRequest

// Mouse parameter structs
type MouseMoveRequest struct {
	Position
} //	@name	MouseMoveRequest

type MouseClickRequest struct {
	Position
	Button string `json:"button"` // left, right, middle
	Double bool   `json:"double"`
} //	@name	MouseClickRequest

type MouseDragRequest struct {
	StartX int    `json:"startX"`
	StartY int    `json:"startY"`
	EndX   int    `json:"endX"`
	EndY   int    `json:"endY"`
	Button string `json:"button"`
} //	@name	MouseDragRequest

type MouseScrollRequest struct {
	Position
	Direction string `json:"direction"` // up, down
	Amount    int    `json:"amount"`
} //	@name	MouseScrollRequest

// Keyboard parameter structs
type KeyboardTypeRequest struct {
	Text  string `json:"text"`
	Delay int    `json:"delay"` // milliseconds between keystrokes
} //	@name	KeyboardTypeRequest

type KeyboardPressRequest struct {
	Key       string   `json:"key"`
	Modifiers []string `json:"modifiers"` // ctrl, alt, shift, cmd
} //	@name	KeyboardPressRequest

type KeyboardHotkeyRequest struct {
	Keys string `json:"keys"` // e.g., "ctrl+c", "cmd+v"
} //	@name	KeyboardHotkeyRequest

// Response structs for keyboard operations
type ScrollResponse struct {
	Success bool `json:"success"`
} //	@name	ScrollResponse

// Response structs
type ScreenshotResponse struct {
	Screenshot     string    `json:"screenshot"`
	CursorPosition *Position `json:"cursorPosition,omitempty"`
	SizeBytes      int       `json:"sizeBytes,omitempty"`
} //	@name	ScreenshotResponse

// Mouse response structs - separated by operation type
type MousePositionResponse struct {
	Position
} //	@name	MousePositionResponse

type MouseClickResponse struct {
	Position
} //	@name	MouseClickResponse

type MouseDragResponse struct {
	Position // Final position
} //	@name	MouseDragResponse

type DisplayInfoResponse struct {
	Displays []DisplayInfo `json:"displays"`
} //	@name	DisplayInfoResponse

type DisplayInfo struct {
	ID int `json:"id"`
	Position
	Size
	IsActive bool `json:"isActive"`
} //	@name	DisplayInfo

type WindowsResponse struct {
	Windows []WindowInfo `json:"windows"`
} //	@name	WindowsResponse

type WindowInfo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Position
	Size
	IsActive bool `json:"isActive"`
} //	@name	WindowInfo

type ComputerUseStatusResponse struct {
	Status string `json:"status"`
} //	@name	ComputerUseStatusResponse

type ComputerUseStartResponse struct {
	Message string                   `json:"message"`
	Status  map[string]ProcessStatus `json:"status"`
} //	@name	ComputerUseStartResponse

type ComputerUseStopResponse struct {
	Message string                   `json:"message"`
	Status  map[string]ProcessStatus `json:"status"`
} //	@name	ComputerUseStopResponse

type ProcessStatus struct {
	Running     bool
	Priority    int
	AutoRestart bool
	Pid         *int
} //	@name	ProcessStatus

type ProcessStatusResponse struct {
	ProcessName string `json:"processName"`
	Running     bool   `json:"running"`
} //	@name	ProcessStatusResponse

type ProcessRestartResponse struct {
	Message     string `json:"message"`
	ProcessName string `json:"processName"`
} //	@name	ProcessRestartResponse

type ProcessLogsResponse struct {
	ProcessName string `json:"processName"`
	Logs        string `json:"logs"`
} //	@name	ProcessLogsResponse

type ProcessErrorsResponse struct {
	ProcessName string `json:"processName"`
	Errors      string `json:"errors"`
} //	@name	ProcessErrorsResponse

type ProcessRequest struct {
	ProcessName string
} //	@name	ProcessRequest

type Empty struct{} //	@name	Empty

func (p *ComputerUsePlugin) Server(*plugin.MuxBroker) (any, error) {
	return &ComputerUseRPCServer{Impl: p.Impl}, nil
}

func (p *ComputerUsePlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (any, error) {
	return &ComputerUseRPCClient{client: c}, nil
}

// TakeScreenshot godoc
//
//	@Summary		Take a screenshot
//	@Description	Take a screenshot of the entire screen
//	@Tags			computer-use
//	@Produce		json
//	@Param			showCursor	query		bool	false	"Whether to show cursor in screenshot"
//	@Success		200			{object}	ScreenshotResponse
//	@Router			/computeruse/screenshot [get]
//
//	@id				TakeScreenshot
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

// TakeRegionScreenshot godoc
//
//	@Summary		Take a region screenshot
//	@Description	Take a screenshot of a specific region of the screen
//	@Tags			computer-use
//	@Produce		json
//	@Param			x			query		int		true	"X coordinate of the region"
//	@Param			y			query		int		true	"Y coordinate of the region"
//	@Param			width		query		int		true	"Width of the region"
//	@Param			height		query		int		true	"Height of the region"
//	@Param			showCursor	query		bool	false	"Whether to show cursor in screenshot"
//	@Success		200			{object}	ScreenshotResponse
//	@Router			/computeruse/screenshot/region [get]
//
//	@id				TakeRegionScreenshot
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

// TakeCompressedScreenshot godoc
//
//	@Summary		Take a compressed screenshot
//	@Description	Take a compressed screenshot of the entire screen
//	@Tags			computer-use
//	@Produce		json
//	@Param			showCursor	query		bool	false	"Whether to show cursor in screenshot"
//	@Param			format		query		string	false	"Image format (png or jpeg)"
//	@Param			quality		query		int		false	"JPEG quality (1-100)"
//	@Param			scale		query		float64	false	"Scale factor (0.1-1.0)"
//	@Success		200			{object}	ScreenshotResponse
//	@Router			/computeruse/screenshot/compressed [get]
//
//	@id				TakeCompressedScreenshot
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

// TakeCompressedRegionScreenshot godoc
//
//	@Summary		Take a compressed region screenshot
//	@Description	Take a compressed screenshot of a specific region of the screen
//	@Tags			computer-use
//	@Produce		json
//	@Param			x			query		int		true	"X coordinate of the region"
//	@Param			y			query		int		true	"Y coordinate of the region"
//	@Param			width		query		int		true	"Width of the region"
//	@Param			height		query		int		true	"Height of the region"
//	@Param			showCursor	query		bool	false	"Whether to show cursor in screenshot"
//	@Param			format		query		string	false	"Image format (png or jpeg)"
//	@Param			quality		query		int		false	"JPEG quality (1-100)"
//	@Param			scale		query		float64	false	"Scale factor (0.1-1.0)"
//	@Success		200			{object}	ScreenshotResponse
//	@Router			/computeruse/screenshot/region/compressed [get]
//
//	@id				TakeCompressedRegionScreenshot
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

// GetMousePosition godoc
//
//	@Summary		Get mouse position
//	@Description	Get the current mouse cursor position
//	@Tags			computer-use
//	@Produce		json
//	@Success		200	{object}	MousePositionResponse
//	@Router			/computeruse/mouse/position [get]
//
//	@id				GetMousePosition
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

// MoveMouse godoc
//
//	@Summary		Move mouse cursor
//	@Description	Move the mouse cursor to the specified coordinates
//	@Tags			computer-use
//	@Accept			json
//	@Produce		json
//	@Param			request	body		MouseMoveRequest	true	"Mouse move request"
//	@Success		200		{object}	MousePositionResponse
//	@Router			/computeruse/mouse/move [post]
//
//	@id				MoveMouse
func WrapMoveMouseHandler(fn func(*MouseMoveRequest) (*MousePositionResponse, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req MouseMoveRequest
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

// Click godoc
//
//	@Summary		Click mouse button
//	@Description	Click the mouse button at the specified coordinates
//	@Tags			computer-use
//	@Accept			json
//	@Produce		json
//	@Param			request	body		MouseClickRequest	true	"Mouse click request"
//	@Success		200		{object}	MouseClickResponse
//	@Router			/computeruse/mouse/click [post]
//
//	@id				Click
func WrapClickHandler(fn func(*MouseClickRequest) (*MouseClickResponse, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req MouseClickRequest
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

// Drag godoc
//
//	@Summary		Drag mouse
//	@Description	Drag the mouse from start to end coordinates
//	@Tags			computer-use
//	@Accept			json
//	@Produce		json
//	@Param			request	body		MouseDragRequest	true	"Mouse drag request"
//	@Success		200		{object}	MouseDragResponse
//	@Router			/computeruse/mouse/drag [post]
//
//	@id				Drag
func WrapDragHandler(fn func(*MouseDragRequest) (*MouseDragResponse, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req MouseDragRequest
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

// Scroll godoc
//
//	@Summary		Scroll mouse wheel
//	@Description	Scroll the mouse wheel at the specified coordinates
//	@Tags			computer-use
//	@Accept			json
//	@Produce		json
//	@Param			request	body		MouseScrollRequest	true	"Mouse scroll request"
//	@Success		200		{object}	ScrollResponse
//	@Router			/computeruse/mouse/scroll [post]
//
//	@id				Scroll
func WrapScrollHandler(fn func(*MouseScrollRequest) (*ScrollResponse, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req MouseScrollRequest
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

// TypeText godoc
//
//	@Summary		Type text
//	@Description	Type text with optional delay between keystrokes
//	@Tags			computer-use
//	@Accept			json
//	@Produce		json
//	@Param			request	body		KeyboardTypeRequest	true	"Text typing request"
//	@Success		200		{object}	Empty
//	@Router			/computeruse/keyboard/type [post]
//
//	@id				TypeText
func WrapTypeTextHandler(fn func(*KeyboardTypeRequest) (*Empty, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req KeyboardTypeRequest
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

// PressKey godoc
//
//	@Summary		Press key
//	@Description	Press a key with optional modifiers
//	@Tags			computer-use
//	@Accept			json
//	@Produce		json
//	@Param			request	body		KeyboardPressRequest	true	"Key press request"
//	@Success		200		{object}	Empty
//	@Router			/computeruse/keyboard/key [post]
//
//	@id				PressKey
func WrapPressKeyHandler(fn func(*KeyboardPressRequest) (*Empty, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req KeyboardPressRequest
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

// PressHotkey godoc
//
//	@Summary		Press hotkey
//	@Description	Press a hotkey combination (e.g., ctrl+c, cmd+v)
//	@Tags			computer-use
//	@Accept			json
//	@Produce		json
//	@Param			request	body		KeyboardHotkeyRequest	true	"Hotkey press request"
//	@Success		200		{object}	Empty
//	@Router			/computeruse/keyboard/hotkey [post]
//
//	@id				PressHotkey
func WrapPressHotkeyHandler(fn func(*KeyboardHotkeyRequest) (*Empty, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req KeyboardHotkeyRequest
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

// GetDisplayInfo godoc
//
//	@Summary		Get display information
//	@Description	Get information about all available displays
//	@Tags			computer-use
//	@Produce		json
//	@Success		200	{object}	DisplayInfoResponse
//	@Router			/computeruse/display/info [get]
//
//	@id				GetDisplayInfo
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

// GetWindows godoc
//
//	@Summary		Get windows information
//	@Description	Get information about all open windows
//	@Tags			computer-use
//	@Produce		json
//	@Success		200	{object}	WindowsResponse
//	@Router			/computeruse/display/windows [get]
//
//	@id				GetWindows
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

// GetStatus godoc
//
//	@Summary		Get computer use status
//	@Description	Get the current status of the computer use system
//	@Tags			computer-use
//	@Produce		json
//	@Success		200	{object}	ComputerUseStatusResponse
//	@Router			/computeruse/status [get]
//
//	@id				GetComputerUseSystemStatus
func WrapStatusHandler(fn func() (*ComputerUseStatusResponse, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		response, err := fn()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	}
}
