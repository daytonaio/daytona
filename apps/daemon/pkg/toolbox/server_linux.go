//go:build linux

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package toolbox

import (
	"log/slog"
	"os"
	"path"

	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/daytonaio/daemon/pkg/toolbox/computeruse/manager"
	recordingcontroller "github.com/daytonaio/daemon/pkg/toolbox/computeruse/recording"
	"github.com/daytonaio/daemon/pkg/toolbox/lsp"
	"github.com/daytonaio/daemon/pkg/toolbox/process/interpreter"
	"github.com/daytonaio/daemon/pkg/toolbox/process/pty"
	"github.com/gin-gonic/gin"
)

var computerUseInstance computeruse.IComputerUse

func (s *server) registerPlatformRoutes(r *gin.Engine) {
	lspLogger := s.logger.With(slog.String("component", "lsp_service"))
	lspController := r.Group("/lsp")
	{
		lspController.POST("/start", lsp.Start(lspLogger))
		lspController.POST("/stop", lsp.Stop(lspLogger))

		lspController.POST("/completions", lsp.Completions(lspLogger))
		lspController.POST("/did-open", lsp.DidOpen(lspLogger))
		lspController.POST("/did-close", lsp.DidClose(lspLogger))

		lspController.GET("/document-symbols", lsp.DocumentSymbols(lspLogger))
		lspController.GET("/workspacesymbols", lsp.WorkspaceSymbols(lspLogger))
	}

	processController := r.Group("/process")
	{
		ptyController := pty.NewPTYController(s.logger, s.WorkDir)
		ptyGroup := processController.Group("/pty")
		{
			ptyGroup.GET("", ptyController.ListPTYSessions)
			ptyGroup.POST("", ptyController.CreatePTYSession)
			ptyGroup.GET("/:sessionId", ptyController.GetPTYSession)
			ptyGroup.DELETE("/:sessionId", ptyController.DeletePTYSession)
			ptyGroup.GET("/:sessionId/connect", ptyController.ConnectPTYSession)
			ptyGroup.POST("/:sessionId/resize", ptyController.ResizePTYSession)
		}

		interpreterController := interpreter.NewInterpreterController(s.logger, s.WorkDir)
		interpreterGroup := processController.Group("/interpreter")
		{
			interpreterGroup.POST("/context", interpreterController.CreateContext)
			interpreterGroup.GET("/context", interpreterController.ListContexts)
			interpreterGroup.DELETE("/context/:id", interpreterController.DeleteContext)
			interpreterGroup.GET("/execute", interpreterController.Execute)
		}
	}

	lazyCU := computeruse.NewLazyComputerUse()
	computerUseInstance = lazyCU

	go func() {
		pluginPath := "/usr/local/lib/daytona-computer-use"
		if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
			pluginPath = path.Join(s.configDir, "daytona-computer-use")
		}

		impl, err := manager.GetComputerUse(s.logger, pluginPath)
		if err != nil {
			s.logger.Error("Computer-Use error", "error", err)
			s.logger.Info("Continuing without computer-use functionality...")
			return
		}
		lazyCU.Set(impl)
		s.logger.Info("Computer-use plugin loaded successfully")
	}()

	computerUseController := r.Group("/computeruse")
	{
		computerUseHandler := computeruse.Handler{
			ComputerUse: lazyCU,
		}

		cuRoutes := computerUseController.Group("/", computeruse.LazyCheckMiddleware(lazyCU))

		cuRoutes.GET("/status", computeruse.WrapStatusHandler(lazyCU.GetStatus))

		cuRoutes.POST("/start", computerUseHandler.StartComputerUse)
		cuRoutes.POST("/stop", computerUseHandler.StopComputerUse)
		cuRoutes.GET("/process-status", computerUseHandler.GetComputerUseStatus)
		cuRoutes.GET("/process/:processName/status", computerUseHandler.GetProcessStatus)
		cuRoutes.POST("/process/:processName/restart", computerUseHandler.RestartProcess)
		cuRoutes.GET("/process/:processName/logs", computerUseHandler.GetProcessLogs)
		cuRoutes.GET("/process/:processName/errors", computerUseHandler.GetProcessErrors)

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

		// Accessibility (AT-SPI) endpoints
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
		s.logger.Info("Stopping computer use...")
		_, err := computerUseInstance.Stop()
		if err != nil {
			s.logger.Error("Failed to stop computer use", "error", err)
		}
	}
}
