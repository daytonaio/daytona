//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package toolbox

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/daytonaio/daemon/pkg/toolbox/computeruse/manager"
	recordingcontroller "github.com/daytonaio/daemon/pkg/toolbox/computeruse/recording"
	"github.com/gin-gonic/gin"
)

var computerUseInstance computeruse.IComputerUse

func resolvePluginPath(configDir string) string {
	const pluginExe = "daytona-computer-use.exe"
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), pluginExe)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	candidate := `C:\ProgramData\Daytona\plugins\` + pluginExe
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	return path.Join(configDir, pluginExe)
}

func (s *server) registerPlatformRoutes(r *gin.Engine) {
	lazyCU := computeruse.NewLazyComputerUse()
	computerUseInstance = lazyCU

	cuHandler := computeruse.Handler{ComputerUse: lazyCU}

	// cuMu serializes the management sections of /start and /stop (spawn+set
	// vs stop+kill+clear) so they cannot interleave: a stop issued during an
	// in-flight spawn waits for the spawn to finish, then kills the fresh
	// instance. manager.GetComputerUse / manager.KillComputerUse additionally
	// hold the manager's own lock, so no code path can spawn twice or leak a
	// child process.
	var cuMu sync.Mutex

	computerUseController := r.Group("/computeruse")
	{
		computerUseController.POST("/start", func(c *gin.Context) {
			cuMu.Lock()
			if !lazyCU.IsReady() {
				pluginPath := resolvePluginPath(s.configDir)
				impl, err := manager.GetComputerUse(s.logger, pluginPath)
				if err != nil {
					cuMu.Unlock()
					c.JSON(http.StatusServiceUnavailable, gin.H{
						"error":   "Failed to spawn computer-use plugin in active console session",
						"details": err.Error(),
						"hint":    "Ensure a user is logged on (AutoLogon) and the plugin binary is at " + pluginPath,
					})
					return
				}
				lazyCU.Set(impl)
				s.logger.Info("Computer-use plugin spawned into active console session", "path", pluginPath)
			}
			cuMu.Unlock()
			cuHandler.StartComputerUse(c)
		})

		computerUseController.POST("/stop", func(c *gin.Context) {
			cuMu.Lock()
			defer cuMu.Unlock()
			if lazyCU.IsReady() {
				cuHandler.StopComputerUse(c)
				manager.KillComputerUse()
				lazyCU.Set(nil)
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Computer-use plugin was not running"})
		})

		cuRoutes := computerUseController.Group("/", computeruse.LazyCheckMiddleware(lazyCU))

		cuRoutes.GET("/status", computeruse.WrapStatusHandler(lazyCU.GetStatus))
		cuRoutes.GET("/process-status", cuHandler.GetComputerUseStatus)
		cuRoutes.GET("/process/:processName/status", cuHandler.GetProcessStatus)
		cuRoutes.POST("/process/:processName/restart", cuHandler.RestartProcess)
		cuRoutes.GET("/process/:processName/logs", cuHandler.GetProcessLogs)
		cuRoutes.GET("/process/:processName/errors", cuHandler.GetProcessErrors)

		cuRoutes.GET("/screenshot", computeruse.WrapScreenshotHandler(lazyCU.TakeScreenshot))
		cuRoutes.GET("/screenshot/region", computeruse.WrapRegionScreenshotHandler(lazyCU.TakeRegionScreenshot))
		cuRoutes.GET("/screenshot/compressed", computeruse.WrapCompressedScreenshotHandler(lazyCU.TakeCompressedScreenshot))
		cuRoutes.GET("/screenshot/region/compressed", computeruse.WrapCompressedRegionScreenshotHandler(lazyCU.TakeCompressedRegionScreenshot))

		cuRoutes.GET("/mouse/position", computeruse.WrapMousePositionHandler(lazyCU.GetMousePosition))
		cuRoutes.POST("/mouse/move", computeruse.WrapMoveMouseHandler(lazyCU.MoveMouse))
		cuRoutes.POST("/mouse/click", computeruse.WrapClickHandler(lazyCU.Click))
		cuRoutes.POST("/mouse/drag", computeruse.WrapDragHandler(lazyCU.Drag))
		cuRoutes.POST("/mouse/scroll", computeruse.WrapScrollHandler(lazyCU.Scroll))

		cuRoutes.POST("/keyboard/type", computeruse.WrapTypeTextHandler(lazyCU.TypeText))
		cuRoutes.POST("/keyboard/key", computeruse.WrapPressKeyHandler(lazyCU.PressKey))
		cuRoutes.POST("/keyboard/hotkey", computeruse.WrapPressHotkeyHandler(lazyCU.PressHotkey))

		cuRoutes.GET("/display/info", computeruse.WrapDisplayInfoHandler(lazyCU.GetDisplayInfo))
		cuRoutes.GET("/display/windows", computeruse.WrapWindowsHandler(lazyCU.GetWindows))

		cuRoutes.GET("/a11y/tree", computeruse.WrapGetAccessibilityTreeHandler(lazyCU.GetAccessibilityTree))
		cuRoutes.POST("/a11y/find", computeruse.WrapFindAccessibilityNodesHandler(lazyCU.FindAccessibilityNodes))
		cuRoutes.POST("/a11y/node/focus", computeruse.WrapFocusAccessibilityNodeHandler(lazyCU.FocusAccessibilityNode))
		cuRoutes.POST("/a11y/node/invoke", computeruse.WrapInvokeAccessibilityNodeHandler(lazyCU.InvokeAccessibilityNode))
		cuRoutes.POST("/a11y/node/value", computeruse.WrapSetAccessibilityNodeValueHandler(lazyCU.SetAccessibilityNodeValue))
	}

	recordingController := recordingcontroller.NewRecordingController(s.recordingService)
	recordingsGroup := computerUseController.Group("/recordings")
	{
		recordingsGroup.POST("/start", recordingController.StartRecording)
		recordingsGroup.POST("/stop", recordingController.StopRecording)
		recordingsGroup.GET("", recordingController.ListRecordings)
		recordingsGroup.GET("/:id", recordingController.GetRecording)
		recordingsGroup.GET("/:id/download", recordingController.DownloadRecording)
		recordingsGroup.DELETE("/:id", recordingController.DeleteRecording)
	}
}

func (s *server) shutdownPlatform() {
	if computerUseInstance != nil {
		s.logger.Info("Stopping computer-use plugin...")
		if _, err := computerUseInstance.Stop(); err != nil {
			s.logger.Error("Failed to stop computer-use plugin", "error", err)
		}
		manager.KillComputerUse()
	}
}
