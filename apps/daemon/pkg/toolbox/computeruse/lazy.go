// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"errors"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

var errNotLoaded = errors.New("computer-use plugin is not loaded yet")

// Compile-time check that LazyComputerUse implements IComputerUse.
var _ IComputerUse = &LazyComputerUse{}

// LazyComputerUse is a proxy that implements IComputerUse and delegates to the
// real implementation once it has been set. Before Set is called, every method
// returns errNotLoaded.
type LazyComputerUse struct {
	mu   sync.RWMutex
	impl IComputerUse
}

func NewLazyComputerUse() *LazyComputerUse {
	return &LazyComputerUse{}
}

// Set stores the real implementation. It is safe to call from any goroutine.
func (l *LazyComputerUse) Set(impl IComputerUse) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.impl = impl
}

// IsReady reports whether the real implementation has been loaded.
func (l *LazyComputerUse) IsReady() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.impl != nil
}

func (l *LazyComputerUse) get() (IComputerUse, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.impl == nil {
		return nil, errNotLoaded
	}
	return l.impl, nil
}

// --- IComputerUse implementation ---

func (l *LazyComputerUse) Initialize() (*Empty, error) {
	impl, err := l.get()
	if err != nil {
		return nil, err
	}
	return impl.Initialize()
}

func (l *LazyComputerUse) Start() (*Empty, error) {
	impl, err := l.get()
	if err != nil {
		return nil, err
	}
	return impl.Start()
}

func (l *LazyComputerUse) Stop() (*Empty, error) {
	impl, err := l.get()
	if err != nil {
		return nil, err
	}
	return impl.Stop()
}

func (l *LazyComputerUse) GetProcessStatus() (map[string]ProcessStatus, error) {
	impl, err := l.get()
	if err != nil {
		return nil, err
	}
	return impl.GetProcessStatus()
}

func (l *LazyComputerUse) IsProcessRunning(req *ProcessRequest) (bool, error) {
	impl, err := l.get()
	if err != nil {
		return false, err
	}
	return impl.IsProcessRunning(req)
}

func (l *LazyComputerUse) RestartProcess(req *ProcessRequest) (*Empty, error) {
	impl, err := l.get()
	if err != nil {
		return nil, err
	}
	return impl.RestartProcess(req)
}

func (l *LazyComputerUse) GetProcessLogs(req *ProcessRequest) (string, error) {
	impl, err := l.get()
	if err != nil {
		return "", err
	}
	return impl.GetProcessLogs(req)
}

func (l *LazyComputerUse) GetProcessErrors(req *ProcessRequest) (string, error) {
	impl, err := l.get()
	if err != nil {
		return "", err
	}
	return impl.GetProcessErrors(req)
}

func (l *LazyComputerUse) TakeScreenshot(req *ScreenshotRequest) (*ScreenshotResponse, error) {
	impl, err := l.get()
	if err != nil {
		return nil, err
	}
	return impl.TakeScreenshot(req)
}

func (l *LazyComputerUse) TakeRegionScreenshot(req *RegionScreenshotRequest) (*ScreenshotResponse, error) {
	impl, err := l.get()
	if err != nil {
		return nil, err
	}
	return impl.TakeRegionScreenshot(req)
}

func (l *LazyComputerUse) TakeCompressedScreenshot(req *CompressedScreenshotRequest) (*ScreenshotResponse, error) {
	impl, err := l.get()
	if err != nil {
		return nil, err
	}
	return impl.TakeCompressedScreenshot(req)
}

func (l *LazyComputerUse) TakeCompressedRegionScreenshot(req *CompressedRegionScreenshotRequest) (*ScreenshotResponse, error) {
	impl, err := l.get()
	if err != nil {
		return nil, err
	}
	return impl.TakeCompressedRegionScreenshot(req)
}

func (l *LazyComputerUse) GetMousePosition() (*MousePositionResponse, error) {
	impl, err := l.get()
	if err != nil {
		return nil, err
	}
	return impl.GetMousePosition()
}

func (l *LazyComputerUse) MoveMouse(req *MouseMoveRequest) (*MousePositionResponse, error) {
	impl, err := l.get()
	if err != nil {
		return nil, err
	}
	return impl.MoveMouse(req)
}

func (l *LazyComputerUse) Click(req *MouseClickRequest) (*MouseClickResponse, error) {
	impl, err := l.get()
	if err != nil {
		return nil, err
	}
	return impl.Click(req)
}

func (l *LazyComputerUse) Drag(req *MouseDragRequest) (*MouseDragResponse, error) {
	impl, err := l.get()
	if err != nil {
		return nil, err
	}
	return impl.Drag(req)
}

func (l *LazyComputerUse) Scroll(req *MouseScrollRequest) (*ScrollResponse, error) {
	impl, err := l.get()
	if err != nil {
		return nil, err
	}
	return impl.Scroll(req)
}

func (l *LazyComputerUse) TypeText(req *KeyboardTypeRequest) (*Empty, error) {
	impl, err := l.get()
	if err != nil {
		return nil, err
	}
	return impl.TypeText(req)
}

func (l *LazyComputerUse) PressKey(req *KeyboardPressRequest) (*Empty, error) {
	impl, err := l.get()
	if err != nil {
		return nil, err
	}
	return impl.PressKey(req)
}

func (l *LazyComputerUse) PressHotkey(req *KeyboardHotkeyRequest) (*Empty, error) {
	impl, err := l.get()
	if err != nil {
		return nil, err
	}
	return impl.PressHotkey(req)
}

func (l *LazyComputerUse) GetDisplayInfo() (*DisplayInfoResponse, error) {
	impl, err := l.get()
	if err != nil {
		return nil, err
	}
	return impl.GetDisplayInfo()
}

func (l *LazyComputerUse) GetWindows() (*WindowsResponse, error) {
	impl, err := l.get()
	if err != nil {
		return nil, err
	}
	return impl.GetWindows()
}

func (l *LazyComputerUse) GetStatus() (*ComputerUseStatusResponse, error) {
	impl, err := l.get()
	if err != nil {
		return nil, err
	}
	return impl.GetStatus()
}

// LazyCheckMiddleware returns 503 if the computer-use plugin has not loaded yet.
func LazyCheckMiddleware(lazy *LazyComputerUse) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !lazy.IsReady() {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"message":  "Computer-use functionality is not available",
				"details":  "The computer-use plugin is still loading or failed to initialize.",
				"solution": "Retry shortly. If the problem persists, check the daemon logs for specific error details.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
